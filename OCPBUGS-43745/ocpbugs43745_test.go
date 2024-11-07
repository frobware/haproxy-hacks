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
	kubeClientset  *kubernetes.Clientset
	kubeConfig     *rest.Config
	logger         *slog.Logger
	routeClientset *routeclientset.Clientset
	backendPods    []*corev1.Pod
	services       []*corev1.Service
	testNamespace  *corev1.Namespace
	testRoute      *routev1.Route
}

type haproxyBackend struct {
	name     string
	settings []string
	servers  []string
}

// parseHAProxyConfig parses the HAProxy configuration content and returns a slice of haproxyBackend.
func parseHAProxyConfig(content string) ([]haproxyBackend, error) {
	var backends []haproxyBackend

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "backend ") {
			backends = append(backends, parseBackend(scanner, line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading HAProxy config: %w", err)
	}

	return backends, nil
}

// parseBackend parses a single backend section in the HAProxy configuration.
func parseBackend(scanner *bufio.Scanner, firstLine string) haproxyBackend {
	backend := haproxyBackend{
		name:     strings.TrimSpace(strings.TrimPrefix(firstLine, "backend")),
		settings: []string{},
		servers:  []string{},
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "backend") {
			// We've reached the next backend
			scanner.Scan() // Move back one line

			break
		}

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		if strings.HasPrefix(trimmedLine, "server") {
			backend.servers = append(backend.servers, trimmedLine)
		} else {
			backend.settings = append(backend.settings, trimmedLine)
		}
	}

	return backend
}

// findBackend searches for a backend with the expected backend and
// server names.
func findBackend(backends []haproxyBackend, expectedBackendName, expectedServiceName string) (haproxyBackend, bool) {
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

// getHAProxyConfig retrieves the HAProxy configuration from the
// router pod.
func (p *routerPod) getHAProxyConfig(ctx context.Context) ([]haproxyBackend, error) {
	stdout, stderr, err := executeCommandWithRetries(ctx, p.kubeClient, p.restConfig, p.name, p.namespace, "router", []string{"cat", "/var/lib/haproxy/conf/haproxy.config"}, 3, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get HAProxy config from pod %s: %w\nstderr: %s", p.name, err, stderr)
	}

	return parseHAProxyConfig(stdout)
}

type waitForAllRoutesAdmittedProgressFunc func(admittedRoutes, totalRoutes int, pendingRoutes []string)

// waitForAllRoutesAdmitted waits until all routes in the namespace are admitted.
func waitForAllRoutesAdmitted(
	routeClient *routeclientset.Clientset,
	namespace string,
	timeout time.Duration,
	progress waitForAllRoutesAdmittedProgressFunc,
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

// getRouterPods retrieves the router pods from the "openshift-ingress" namespace.
func getRouterPods(kubeClient *kubernetes.Clientset, restConfig *rest.Config) ([]*routerPod, error) {
	pods, err := kubeClient.CoreV1().Pods("openshift-ingress").List(context.Background(), metav1.ListOptions{
		LabelSelector: "ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list router pods: %w", err)
	}

	routerPods := make([]*routerPod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		routerPods = append(routerPods, &routerPod{
			name:       pod.Name,
			namespace:  pod.Namespace,
			kubeClient: kubeClient,
			restConfig: restConfig,
		})
	}

	return routerPods, nil
}

// waitForHAProxyConfigCondition waits until the HAProxy configuration meets the expected condition.
func waitForHAProxyConfigCondition(
	ctx context.Context,
	routerPods []*routerPod,
	expectedBackendName, expectedServerName string,
	shouldBePresent bool,
	logger *slog.Logger,
) error {
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		for _, routerPod := range routerPods {
			backends, err := routerPod.getHAProxyConfig(ctx)
			if err != nil {
				return false, err
			}

			backend, found := findBackend(backends, expectedBackendName, expectedServerName)

			if found == shouldBePresent {
				if found {
					logger.Info("HAProxy backend entry found",
						"pod", routerPod.name,
						"backend", expectedBackendName,
						"server", expectedServerName,
						"config", fmt.Sprintf("backend %s\n  %s\n  %s",
							backend.name,
							strings.Join(backend.settings, "\n  "),
							strings.Join(backend.servers, "\n  ")))
				} else {
					logger.Info("Backend entry absent as expected",
						"pod", routerPod.name,
						"route", expectedBackendName)
				}
			} else {
				logger.Info("HAProxy backend entry not found",
					"pod", routerPod.name,
					"backend", expectedBackendName,
					"server", expectedServerName,
					"shouldBePresent", shouldBePresent,
					"found", found)

				return false, nil
			}
		}

		return true, nil
	})
}

// waitForHAProxyConfigUpdate waits for the HAProxy configuration to update after switching services.
func waitForHAProxyConfigUpdate(
	ctx context.Context,
	kubeClient *kubernetes.Clientset,
	restConfig *rest.Config,
	route *routev1.Route,
	service *corev1.Service,
	backendPod *corev1.Pod,
	logger *slog.Logger,
) error {
	logger.Info("Waiting for HAProxy configuration update", "service", service.Name)

	// Get the router pods
	routerPods, err := getRouterPods(kubeClient, restConfig)
	if err != nil {
		return fmt.Errorf("failed to get router pods: %w", err)
	}

	// Construct the expected backend and service names
	expectedBackendName := fmt.Sprintf("be_http:%s:%s", route.Namespace, route.Name)
	expectedServerName := fmt.Sprintf("pod:%s:%s", backendPod.Name, service.Name)

	// Wait for the HAProxy configuration condition to be met
	err = waitForHAProxyConfigCondition(ctx, routerPods, expectedBackendName, expectedServerName, true, logger)
	if err != nil {
		return fmt.Errorf("failed waiting for HAProxy configuration update: %w", err)
	}

	logger.Info("HAProxy configuration updated successfully", "backend", expectedBackendName, "server", expectedServerName)

	return nil
}

// createHTTPClient creates an HTTP client with the specified timeout.
func createHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// constructRequest constructs an HTTP request with the given method and URL.
func constructRequest(ctx context.Context, method, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, nil)
}

// performRequest executes the HTTP request and returns the response.
func performRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

// readResponseBody reads the response body from the HTTP response.
func readResponseBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getRouteResponse sends a GET request to the route's host and returns the response body.
func getRouteResponse(logger *slog.Logger, route *routev1.Route) (string, error) {
	if route.Spec.Host == "" {
		return "", fmt.Errorf("route %s/%s has no host", route.Namespace, route.Name)
	}

	url := "http://" + route.Spec.Host
	logger.Info("Making GET request", "url", url)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := constructRequest(ctx, http.MethodGet, url)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := createHTTPClient(10 * time.Second)

	resp, err := performRequest(client, req)
	if err != nil {
		return "", fmt.Errorf("GET request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := readResponseBody(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// fetchServiceResponse retrieves the response from the service behind the route.
func fetchServiceResponse(logger *slog.Logger, route *routev1.Route) (string, error) {
	logger.Info("Getting response from service", "service", route.Spec.To.Name)
	response, err := getRouteResponse(logger, route)

	if err != nil {
		return "", fmt.Errorf("failed to get response from service: %w", err)
	}

	logger.Info("Received response from service", "service", route.Spec.To.Name, "response", response)

	return response, nil
}

// updateRouteService updates the route to point to a new service.
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

// routeSwitchServiceAndVerifyResponse switches the route to point to a new service and verifies the HAProxy configuration.
func routeSwitchServiceAndVerifyResponse(
	ctx context.Context,
	kubeClient *kubernetes.Clientset,
	restConfig *rest.Config,
	routeClient *routeclientset.Clientset,
	route *routev1.Route,
	service *corev1.Service,
	pod *corev1.Pod,
	logger *slog.Logger,
) (*routev1.Route, error) {
	updatedRoute, err := updateRouteService(ctx, routeClient, route, service.Name, logger)
	if err != nil {
		return nil, err
	}

	if err := waitForHAProxyConfigUpdate(ctx, kubeClient, restConfig, updatedRoute, service, pod, logger); err != nil {
		return nil, err
	}

	return updatedRoute, nil
}

// switchRouteServiceAndFetchResponse switches the route to a specified service index and fetches the response.
func switchRouteServiceAndFetchResponse(
	ctx context.Context,
	tc *TestConfig,
	serviceIndex int,
	delay time.Duration,
) (string, error) {
	service := tc.services[serviceIndex]
	backendPod := tc.backendPods[serviceIndex]

	updatedRoute, err := routeSwitchServiceAndVerifyResponse(
		ctx,
		tc.kubeClientset,
		tc.kubeConfig,
		tc.routeClientset,
		tc.testRoute,
		service,
		backendPod,
		tc.logger,
	)
	if err != nil {
		return "", fmt.Errorf("failed during switch to service %s: %w", service.Name, err)
	}

	tc.testRoute = updatedRoute
	tc.logger.Info("Switched to service", "service", service.Name)

	if _, err := waitForAllRoutesAdmitted(tc.routeClientset, tc.testNamespace.Name, 30*time.Second, func(admittedRoutes, totalRoutes int, pendingRoutes []string) {
		if len(pendingRoutes) > 0 {
			tc.logger.Info("Not all routes have been admitted", "admittedRoutes", admittedRoutes, "totalRoutes", totalRoutes, "pending", strings.Join(pendingRoutes, ", "))
		}
	}); err != nil {
		return "", fmt.Errorf("not all routes have been admitted in namespace %s: %w", tc.testNamespace.Name, err)
	}

	tc.logger.Info("All routes admitted", "namespace", tc.testNamespace.Name)
	tc.logger.Info("Delaying GET request", "route", tc.testRoute.Name, "duration", delay)
	time.Sleep(delay)

	return fetchServiceResponse(tc.logger, tc.testRoute)
}

// waitForReplicationControllerReady waits for the replication controller to have the desired number of ready replicas.
func waitForReplicationControllerReady(ctx context.Context, kubeClient *kubernetes.Clientset, rc *corev1.ReplicationController, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		current, err := kubeClient.CoreV1().ReplicationControllers(rc.Namespace).Get(ctx, rc.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return current.Status.ReadyReplicas >= *current.Spec.Replicas, nil
	})
}

// createRoute creates a new route pointing to the specified service.
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

// RetryWithDelay retries a given operation with a delay between
// attempts.
func RetryWithDelay(ctx context.Context, attempts int, delay time.Duration, operation func() error) error {
	for i := 0; i < attempts; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		} else {
			return err
		}
	}

	return nil
}

// executeCommandWithRetries executes a command in a pod with retries.
func executeCommandWithRetries(ctx context.Context, kubeClient *kubernetes.Clientset, restConfig *rest.Config, podName, namespace, container string, command []string, attempts int, delay time.Duration) (string, string, error) {
	var stdout, stderr string

	err := RetryWithDelay(ctx, attempts, delay, func() error {
		var err error
		stdout, stderr, err = executeCommandInPod(ctx, kubeClient, restConfig, podName, namespace, container, command)

		return err
	})

	return stdout, stderr, err
}

// executeCommandInPod executes a command in a specific pod container.
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

// TestRouteServiceSwitch tests switching services behind a route and verifies the response changes.
func TestRouteServiceSwitch(t *testing.T) {
	var getResponseDelay time.Duration

	if v := os.Getenv("REQUEST_DELAY"); v != "" {
		n, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("Invalid duration: %s: %v", v, err)
		}

		getResponseDelay = n
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	tc := &TestConfig{}
	{
		logger := slog.Default()

		cfg, err := config.GetConfig()
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}

		kubeClientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("failed to create kubernetes client: %v", err)
		}

		routeClientset, err := routeclientset.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("failed to create route client: %v", err)
		}

		configClientset, err := configclientset.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("failed to create config clientset: %v", err)
		}

		clusterVersion, err := configClientset.ConfigV1().ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("failed to retrieve cluster version: %v", err)
		}

		logger.Info("Running test on OpenShift Cluster Version", "version", clusterVersion.Status.Desired.Version)

		tc.logger = logger
		tc.kubeConfig = cfg
		tc.kubeClientset = kubeClientset
		tc.routeClientset = routeClientset

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "route-service-switcher-test-",
				Labels:       map[string]string{"test": "namespace"},
			},
		}

		ns, err = kubeClientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create namespace: %v", err)
		}

		tc.testNamespace = ns

		int32Ptr := func(i int32) *int32 { return &i }

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

			rc, err = kubeClientset.CoreV1().ReplicationControllers(ns.Name).Create(ctx, rc, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("failed to create ReplicationController %d: %v", i, err)
			}

			err = waitForReplicationControllerReady(ctx, kubeClientset, rc, 2*time.Minute)
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

			svc, err = kubeClientset.CoreV1().Services(ns.Name).Create(ctx, svc, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("failed to create Service %d: %v", i, err)
			}

			tc.services = append(tc.services, svc)

			podList, err := kubeClientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=web-server,instance=%d", i),
			})
			if err != nil {
				t.Fatalf("failed to list pods for RC %d: %v", i, err)
			}

			if len(podList.Items) == 0 {
				t.Fatalf("no pods found for RC %d", i)
			}

			if len(podList.Items) != 1 {
				t.Fatalf("only expected 1 pod found for RC, got %d", i)
			}

			tc.backendPods = append(tc.backendPods, &podList.Items[0])

			_, err = createRoute(ctx, tc.routeClientset, ns.Name, "svc-"+strconv.Itoa(i), svc.Name, nil)
			if err != nil {
				t.Fatalf("Failed to create route: %v", err)
			}
		}

		route, err := createRoute(ctx, tc.routeClientset, ns.Name, "test", tc.services[0].Name, nil)
		if err != nil {
			t.Fatalf("Failed to create test route: %v", err)
		}

		tc.testRoute = route
	}

	if v := os.Getenv("NO_CLEANUP"); v != "1" {
		t.Cleanup(func() {
			tc.kubeClientset.CoreV1().Namespaces().Delete(context.Background(), tc.testNamespace.Name, metav1.DeleteOptions{})
		})
	}

	t.Run("switching between services returns different responses", func(t *testing.T) {
		if len(tc.services) < 2 || len(tc.backendPods) < 2 {
			t.Fatal("Not enough services or pods to test switching")
		}

		resp1, err := switchRouteServiceAndFetchResponse(ctx, tc, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		resp2, err := switchRouteServiceAndFetchResponse(ctx, tc, 1, getResponseDelay)
		if err != nil {
			t.Fatal(err)
		}

		if resp1 == resp2 {
			t.Fatalf("Expected different responses after switching services, but got the same response: %s", resp1)
		}
	})
}
