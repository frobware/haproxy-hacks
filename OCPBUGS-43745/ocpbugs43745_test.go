package ocpbugs43745_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	configclientset "github.com/openshift/client-go/config/clientset/versioned"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestConfig struct {
	Logger      *slog.Logger
	KubeConfig  *rest.Config
	KubeClient  *kubernetes.Clientset
	RouteClient *routeclientset.Clientset
	Namespace   *corev1.Namespace
	Services    []*corev1.Service
	Route       *routev1.Route
	Pods        []corev1.Pod
}

type haproxyBackend struct {
	name     string
	settings []string
	servers  []string
}

func parseHAProxyConfig(content string) []haproxyBackend {
	var (
		backends []haproxyBackend
		current  *haproxyBackend
	)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "backend") {
			if current != nil {
				backends = append(backends, *current)
			}

			current = &haproxyBackend{
				name:    strings.TrimSpace(strings.TrimPrefix(line, "backend")),
				servers: []string{},
			}

			continue
		}

		if current != nil && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "server") {
				current.servers = append(current.servers, trimmedLine)
			} else {
				current.settings = append(current.settings, trimmedLine)
			}
		}
	}

	if current != nil {
		backends = append(backends, *current)
	}

	return backends
}

func findBackend(backends []haproxyBackend, route *routev1.Route, service *corev1.Service, pod *corev1.Pod) (haproxyBackend, bool) {
	expectedBackendName := fmt.Sprintf("be_http:%s:%s", route.Namespace, route.Name)
	expectedServiceName := fmt.Sprintf("pod:%s:%s", pod.Name, service.Name)

	for _, b := range backends {
		if b.name == expectedBackendName {
			for _, server := range b.servers {
				if strings.Contains(server, expectedServiceName) {
					return b, true
				}
			}
		}
	}

	return haproxyBackend{}, false
}

type routerPod struct {
	name       string
	namespace  string
	kubeClient *kubernetes.Clientset
	restConfig *rest.Config
}

func (p *routerPod) getHAProxyConfig(ctx context.Context) ([]haproxyBackend, error) {
	stdout, stderr, err := executeCommandWithRetries(ctx, p.kubeClient, p.restConfig, p.name, p.namespace, "router", []string{"cat", "/var/lib/haproxy/conf/haproxy.config"}, 3, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get HAProxy config from pod %s: %w\nstderr: %s", p.name, err, stderr)
	}

	return parseHAProxyConfig(stdout), nil
}

type waitForAllRoutesAddmittedProgressFunc func(admittedRoutes, totalRoutes int, pendingRoutes []string)

func waitForAllRoutesAdmitted(
	routeClient *routeclientset.Clientset,
	namespace string,
	timeout time.Duration,
	progress waitForAllRoutesAddmittedProgressFunc,
) (*routev1.RouteList, error) {
	isRouteAdmitted := func(route *routev1.Route) bool {
		for _, ingress := range route.Status.Ingress {
			if ingress.RouterCanonicalHostname != "" {
				return true
			}
		}

		return false
	}

	var routeList *routev1.RouteList

	err := wait.PollUntilContextTimeout(context.Background(), time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		var err error

		routeList, err = routeClient.RouteV1().Routes(namespace).List(ctx, metav1.ListOptions{})

		if err != nil {
			return false, fmt.Errorf("failed to list routes in namespace %s: %w", namespace, err)
		}

		admittedRoutes := 0

		var pendingRoutes []string

		for _, route := range routeList.Items {
			if isRouteAdmitted(&route) {
				admittedRoutes++
			} else {
				pendingRoutes = append(pendingRoutes, fmt.Sprintf("%s/%s", route.Namespace, route.Name))
			}
		}

		totalRoutes := len(routeList.Items)
		if progress != nil {
			progress(admittedRoutes, totalRoutes, pendingRoutes)
		}

		if admittedRoutes == totalRoutes {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("not all routes were admitted in namespace %s: %w", namespace, err)
	}

	return routeList, nil
}

func waitForHAProxyConfigCondition(tc *TestConfig, routeName string, service *corev1.Service, servicePod *corev1.Pod, shouldBePresent bool) error {
	pods, err := tc.KubeClient.CoreV1().Pods("openshift-ingress").List(context.Background(), metav1.ListOptions{
		LabelSelector: "ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default",
	})
	if err != nil {
		return fmt.Errorf("failed to list router pods: %w", err)
	}

	routerPods := make([]*routerPod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		routerPods = append(routerPods, &routerPod{
			name:       pod.Name,
			namespace:  pod.Namespace,
			kubeClient: tc.KubeClient,
			restConfig: tc.KubeConfig,
		})
	}

	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		for _, pod := range routerPods {
			backends, err := pod.getHAProxyConfig(ctx)
			if err != nil {
				return false, err
			}

			backend, found := findBackend(backends, tc.Route, service, servicePod)
			if found == shouldBePresent {
				if found {
					tc.Logger.Info("backend entry found",
						"pod", pod.name,
						"route", routeName,
						"config", fmt.Sprintf("backend %s\n  %s\n  %s",
							backend.name,
							strings.Join(backend.settings, "\n  "),
							strings.Join(backend.servers, "\n  ")))
				} else {
					tc.Logger.Info("backend entry absent as expected",
						"pod", pod.name,
						"route", routeName)
				}
			} else {
				tc.Logger.Info("config check failed",
					"pod", pod.name,
					"route", routeName,
					"service", service.Name,
					"shouldBePresent", shouldBePresent,
					"found", found)

				return false, nil
			}
		}

		return true, nil
	})
}

func waitForHAProxyConfigUpdate(tc *TestConfig, routeName string, service *corev1.Service, pod *corev1.Pod) error {
	tc.Logger.Info("waiting for HAProxy configuration update", "service", service.Name)
	err := waitForHAProxyConfigCondition(tc, routeName, service, pod, true)

	if err != nil {
		return fmt.Errorf("failed waiting for HAProxy configuration update: %w", err)
	}

	tc.Logger.Info("HAProxy configuration updated successfully", "service", service.Name)

	return nil
}

func getRouteResponse(logger *slog.Logger, route *routev1.Route) (string, error) {
	if route.Spec.Host == "" {
		return "", fmt.Errorf("route %s/%s has no host", route.Namespace, route.Name)
	}

	url := "http://" + route.Spec.Host
	logger.Info("making GET request", "url", url)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func fetchServiceResponse(tc *TestConfig, route *routev1.Route) (string, error) {
	tc.Logger.Info("getting response from service", "service", route.Spec.To.Name)
	response, err := getRouteResponse(tc.Logger, route)

	if err != nil {
		return "", fmt.Errorf("failed to get response from service: %w", err)
	}

	tc.Logger.Info("received response from service", "service", route.Spec.To.Name, "response", response)

	return response, nil
}

func updateRouteService(ctx context.Context, routeClient *routeclientset.Clientset, route *routev1.Route, newServiceName string, logger *slog.Logger) (*routev1.Route, error) {
	var updatedRoute *routev1.Route

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentRoute, err := routeClient.RouteV1().Routes(route.Namespace).Get(ctx, route.Name, metav1.GetOptions{})
		if err != nil {
			logger.Error("Failed to get the latest version of the route", "route", route.Name, "error", err)
			return err
		}

		currentRoute.Spec.To.Name = newServiceName

		updatedRoute, err = routeClient.RouteV1().Routes(route.Namespace).Update(ctx, currentRoute, metav1.UpdateOptions{})
		if err != nil {
			logger.Error("Failed to update the route", "route", route.Name, "error", err)
		} else {
			logger.Info("Successfully updated the route", "route", route.Name, "newServiceName", newServiceName)
		}

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update route %s: %w", route.Name, err)
	}

	return updatedRoute, nil
}

func routeSwitchServiceAndVerifyResponse(tc *TestConfig, service *corev1.Service, pod *corev1.Pod) error {
	updatedRoute, err := updateRouteService(context.Background(), tc.RouteClient, tc.Route, service.Name, tc.Logger)
	if err != nil {
		return err
	}

	tc.Route = updatedRoute

	if err := waitForHAProxyConfigUpdate(tc, updatedRoute.Name, service, pod); err != nil {
		return err
	}

	return nil
}

func switchRouteServiceAndFetchResponse(
	tc *TestConfig,
	serviceIndex int,
	delay time.Duration,
) (string, error) {
	if err := routeSwitchServiceAndVerifyResponse(tc, tc.Services[serviceIndex], &tc.Pods[serviceIndex]); err != nil {
		return "", fmt.Errorf("failed during switch to service %s: %w", tc.Services[serviceIndex].Name, err)
	}

	tc.Logger.Info("Switched to service", "service", tc.Services[serviceIndex].Name)

	if _, err := waitForAllRoutesAdmitted(tc.RouteClient, tc.Namespace.Name, 30*time.Second, func(admittedRoutes, totalRoutes int, pendingRoutes []string) {
		if len(pendingRoutes) > 0 {
			tc.Logger.Info("Pending routes", "admittedRoutes", admittedRoutes, "totalRoutes", totalRoutes, "routes", strings.Join(pendingRoutes, ", "))
		}
	}); err != nil {
		return "", fmt.Errorf("not all routes have been admitted in namespace %s: %w", tc.Namespace.Name, err)
	}

	tc.Logger.Info("All routes admitted", "namespace", tc.Namespace.Name)
	tc.Logger.Info("Delaying GET request", "route", tc.Route.Name, "duation", delay)
	time.Sleep(delay)

	return fetchServiceResponse(tc, tc.Route)
}

func waitForReplicationControllerReady(ctx context.Context, kubeClient *kubernetes.Clientset, rc *corev1.ReplicationController, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		current, err := kubeClient.CoreV1().ReplicationControllers(rc.Namespace).Get(ctx, rc.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return current.Status.ReadyReplicas >= *current.Spec.Replicas, nil
	})
}

func createRoute(ctx context.Context, routeClient *routeclientset.Clientset, namespace, name, serviceName string, labels map[string]string) (*routev1.Route, error) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: serviceName,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}

	return routeClient.RouteV1().Routes(namespace).Create(ctx, route, metav1.CreateOptions{})
}

func executeCommandWithRetries(ctx context.Context, kubeClient *kubernetes.Clientset, restConfig *rest.Config, podName, namespace, container string, command []string, attempts int, delay time.Duration) (string, string, error) {
	var (
		stdout string
		stderr string
		err    error
	)

	for i := 0; i < attempts; i++ {
		stdout, stderr, err = executeCommandInPod(ctx, kubeClient, restConfig, podName, namespace, container, command)
		if err == nil {
			return stdout, stderr, nil
		}

		if i >= attempts-1 {
			return stdout, stderr, err
		}

		select {
		case <-ctx.Done():
			return stdout, stderr, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return "", "", err
}

func executeCommandInPod(ctx context.Context, kubeClient *kubernetes.Clientset, restConfig *rest.Config, podName, namespace, container string, command []string) (string, string, error) {
	req := kubeClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to execute command: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}

func setupTest(t *testing.T) *TestConfig {
	int32Ptr := func(i int32) *int32 { return &i }

	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	logger := slog.Default()

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create kubernetes client: %v", err)
	}

	routeClient, err := routeclientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create route client: %v", err)
	}

	configClient, err := configclientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create config clientset: %v", err)
	}

	clusterVersion, err := configClient.ConfigV1().ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to retrieve cluster version: %v", err)
	}

	logger.Info("Running test on OpenShift Cluster Version", "version", clusterVersion.Status.Desired.Version)

	tc := &TestConfig{
		Logger:      logger,
		KubeConfig:  cfg,
		KubeClient:  kubeClient,
		RouteClient: routeClient,
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "route-service-switcher-test-",
			Labels:       map[string]string{"test": "namespace"},
		},
	}

	ns, err = kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}

	tc.Namespace = ns

	for i := 1; i <= 2; i++ {
		instanceLabel := strconv.Itoa(i)
		labels := map[string]string{"app": "web-server", "instance": instanceLabel}

		rc := &corev1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("web-server-%d", i),
				Namespace: ns.Name,
				Labels:    labels,
			},
			Spec: corev1.ReplicationControllerSpec{
				Replicas: int32Ptr(1),
				Selector: labels,
				Template: &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "quay.io/openshifttest/nginx-alpine@sha256:04f316442d48ba60e3ea0b5a67eb89b0b667abf1c198a3d0056ca748736336a0",
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										Protocol:      corev1.ProtocolTCP,
										ContainerPort: 8080,
									},
								},
							},
						},
					},
				},
			},
		}

		rc, err = kubeClient.CoreV1().ReplicationControllers(ns.Name).Create(ctx, rc, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create ReplicationController %d: %v", i, err)
		}

		err = waitForReplicationControllerReady(ctx, kubeClient, rc, 2*time.Minute)
		if err != nil {
			t.Fatalf("ReplicationController %d is not ready: %v", i, err)
		}

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("service-%d", i),
				Namespace: ns.Name,
				Labels:    labels,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       8080,
						Protocol:   corev1.ProtocolTCP,
						TargetPort: intstr.FromInt(8080),
					},
				},
				Selector: labels,
			},
		}

		svc, err = kubeClient.CoreV1().Services(ns.Name).Create(ctx, svc, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create Service %d: %v", i, err)
		}

		tc.Services = append(tc.Services, svc)

		podList, err := kubeClient.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=web-server,instance=%d", i),
		})
		if err != nil {
			t.Fatalf("failed to list pods for RC %d: %v", i, err)
		}

		if len(podList.Items) == 0 {
			t.Fatalf("no pods found for RC %d", i)
		}

		tc.Pods = append(tc.Pods, podList.Items[0])

		_, err = createRoute(context.Background(), tc.RouteClient, ns.Name, "svc"+strconv.Itoa(i), svc.Name, nil)
		if err != nil {
			t.Fatalf("Failed to create test-route: %v", err)
		}
	}

	route, err := createRoute(context.Background(), tc.RouteClient, ns.Name, "test", tc.Services[0].Name, nil)
	if err != nil {
		t.Fatalf("Failed to create test: %v", err)
	}
	tc.Route = route

	return tc
}

func TestRouteServiceSwitch(t *testing.T) {
	var delay time.Duration

	if v := os.Getenv("SWITCH_DELAY"); v != "" {
		n, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("Invalid duration: %s: %v", v, err)
		}

		delay = n
	}

	tc := setupTest(t)

	if v := os.Getenv("NO_CLEANUP"); v != "1" {
		t.Cleanup(func() {
			tc.KubeClient.CoreV1().Namespaces().Delete(context.Background(), tc.Namespace.Name, metav1.DeleteOptions{})
		})
	}

	t.Run("switching between services returns different responses", func(t *testing.T) {
		if len(tc.Services) < 2 || len(tc.Pods) < 2 {
			t.Fatal("Not enough services or pods to test switching")
		}

		resp1, err := switchRouteServiceAndFetchResponse(tc, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		resp2, err := switchRouteServiceAndFetchResponse(tc, 1, delay)
		if err != nil {
			t.Fatal(err)
		}

		if resp1 == resp2 {
			t.Fatalf("Expected different responses after switching services, but got the same response: %s", resp1)
		}
	})
}
