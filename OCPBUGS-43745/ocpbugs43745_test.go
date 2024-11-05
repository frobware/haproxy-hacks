package ocpbugs43745_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	routev1 "github.com/openshift/api/route/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"

	ocpbugs43745 "github.com/frobware/haproxy-hacks/OCPBUGS-43745"
)

type TestConfig struct {
	Context     context.Context
	Logger      *slog.Logger
	KubeConfig  *rest.Config
	KubeClient  *kubernetes.Clientset
	RouteClient *routeclientset.Clientset
	Namespace   *corev1.Namespace
	Services    []*corev1.Service
	Route       *routev1.Route
	Pods        []corev1.Pod
	PodExecutor ocpbugs43745.PodExecutor
	Cleaner     *ocpbugs43745.Cleaner
	cancel      context.CancelFunc
}

type callerInfo struct {
	file string
	line int
}

func getCallerInfo() callerInfo {
	_, fullPath, line, ok := runtime.Caller(4)
	if !ok {
		return callerInfo{file: "unknown", line: 0}
	}
	return callerInfo{
		file: filepath.Base(fullPath),
		line: line,
	}
}

type slogt struct {
	t     *testing.T
	attrs []slog.Attr
	group string
}

func (h *slogt) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *slogt) Handle(ctx context.Context, r slog.Record) error {
	caller := getCallerInfo()
	attrs := h.formatAttributes(r)
	msg := h.formatMessage(r.Message, attrs)

	log.Printf("    %s:%d: %s\n", caller.file, caller.line, msg)
	return nil
}

func (h *slogt) formatAttributes(r slog.Record) []string {
	attrs := make([]string, 0)

	// Add existing attributes.
	for _, a := range h.attrs {
		attrs = append(attrs, formatAttr(a))
	}

	// Add group if present.
	if h.group != "" {
		attrs = append([]string{formatAttr(slog.String("group", h.group))}, attrs...)
	}

	// Add record attributes.
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, formatAttr(a))
		return true
	})

	return attrs
}

func formatAttr(a slog.Attr) string {
	return fmt.Sprintf("%s=%s", a.Key, strings.TrimSpace(fmt.Sprint(a.Value)))
}

func (h *slogt) formatMessage(msg string, attrs []string) string {
	if len(attrs) > 0 {
		return fmt.Sprintf("%s (%s)", msg, strings.Join(attrs, " "))
	}
	return msg
}

func (h *slogt) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := append(h.attrs, attrs...)
	return &slogt{t: h.t, attrs: newAttrs}
}

func (h *slogt) WithGroup(name string) slog.Handler {
	return &slogt{
		t:     h.t,
		attrs: h.attrs,
		group: name,
	}
}

type haproxyBackend struct {
	name     string
	settings []string
	servers  []string
}

func parseHAProxyConfig(content string) []haproxyBackend {
	var backends []haproxyBackend
	var current *haproxyBackend

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
	name      string
	namespace string
	executor  ocpbugs43745.PodExecutor
}

func (p *routerPod) getHAProxyConfig(ctx context.Context) ([]haproxyBackend, error) {
	stdout, stderr, err := p.executor.Execute(ctx, p.name, p.namespace, "router", []string{"cat", "/var/lib/haproxy/conf/haproxy.config"})
	if err != nil {
		return nil, fmt.Errorf("failed to get HAProxy config from pod %s: %v\nstderr: %s", p.name, err, stderr)
	}
	return parseHAProxyConfig(stdout), nil
}

type configPoller struct {
	interval time.Duration
	timeout  time.Duration
	logger   *slog.Logger
}

func newConfigPoller(interval, timeout time.Duration, logger *slog.Logger) *configPoller {
	return &configPoller{
		interval: interval,
		timeout:  timeout,
		logger:   logger,
	}
}

func (p *configPoller) waitForCondition(ctx context.Context, check func(context.Context) (bool, error)) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for condition: %w", ctx.Err())
		case <-ticker.C:
			ok, err := check(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// Resource creators
type rcCreator struct {
	clientset *kubernetes.Clientset
}

func (r *rcCreator) Create(ctx context.Context, meta ocpbugs43745.ResourceMeta) (*corev1.ReplicationController, error) {
	rc := corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Labels:    meta.Labels,
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: ocpbugs43745.Int32Ptr(1),
			Selector: meta.Labels,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: meta.Labels,
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
	return r.clientset.CoreV1().ReplicationControllers(meta.Namespace).Create(ctx, &rc, metav1.CreateOptions{})
}

type serviceCreator struct {
	clientset *kubernetes.Clientset
}

func (s *serviceCreator) Create(ctx context.Context, meta ocpbugs43745.ResourceMeta) (*corev1.Service, error) {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Labels:    meta.Labels,
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
			Selector: meta.Labels,
		},
	}
	return s.clientset.CoreV1().Services(meta.Namespace).Create(ctx, &svc, metav1.CreateOptions{})
}

type routeCreator struct {
	routeClient *routeclientset.Clientset
}

func (r *routeCreator) Create(ctx context.Context, meta ocpbugs43745.ResourceMeta) (*routev1.Route, error) {
	route := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Labels:    meta.Labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("service-insecure%d", 1),
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
	return r.routeClient.RouteV1().Routes(meta.Namespace).Create(ctx, &route, metav1.CreateOptions{})
}

func setupTest(t *testing.T) *TestConfig {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	logger := slog.New(&slogt{t: t})

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

	tc := &TestConfig{
		Context:     ctx,
		Logger:      logger,
		KubeConfig:  cfg,
		KubeClient:  kubeClient,
		RouteClient: routeClient,
		Cleaner:     ocpbugs43745.NewCleaner(),
		cancel:      cancel,
	}

	nsCreator := ocpbugs43745.NewLoggingCreator(&ocpbugs43745.NamespaceCreator{ClientSet: kubeClient}, logger)
	ns, err := nsCreator.Create(ctx, ocpbugs43745.ResourceMeta{
		Name:   "route-switcher-test-",
		Labels: map[string]string{"test": "namespace"},
	})
	if err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	tc.Namespace = ns
	tc.Cleaner.Add(func() error {
		return kubeClient.CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{})
	})

	rcCheck := func(ctx context.Context, rc *corev1.ReplicationController) (bool, error) {
		current, err := kubeClient.CoreV1().ReplicationControllers(rc.Namespace).Get(ctx, rc.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return current.Status.ReadyReplicas >= *current.Spec.Replicas, nil
	}

	for i := 1; i <= 2; i++ {
		rcCreator := ocpbugs43745.NewReadinessAwareCreator(
			ocpbugs43745.NewLoggingCreator(&rcCreator{clientset: kubeClient}, logger),
			rcCheck, 2*time.Minute,
		)

		_, err := rcCreator.Create(ctx, ocpbugs43745.ResourceMeta{
			Name:      fmt.Sprintf("web-server-rc%d", i),
			Namespace: ns.Name,
			Labels:    map[string]string{"app": "web-server", "instance": fmt.Sprintf("%d", i)},
		})
		if err != nil {
			t.Fatalf("failed to create RC %d: %v", i, err)
		}

		svcCreator := ocpbugs43745.NewLoggingCreator(&serviceCreator{clientset: kubeClient}, logger)
		svc, err := svcCreator.Create(ctx, ocpbugs43745.ResourceMeta{
			Name:      fmt.Sprintf("service-insecure%d", i),
			Namespace: ns.Name,
			Labels:    map[string]string{"app": "web-server", "instance": fmt.Sprintf("%d", i)},
		})
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
		tc.Pods = append(tc.Pods, podList.Items...)
	}

	routeCreator := ocpbugs43745.NewLoggingCreator(&routeCreator{routeClient: routeClient}, logger)
	route, err := routeCreator.Create(ctx, ocpbugs43745.ResourceMeta{
		Name:      "test-route",
		Namespace: ns.Name,
		Labels:    map[string]string{"app": "web-server"},
	})
	if err != nil {
		t.Fatalf("failed to create route: %v", err)
	}
	tc.Route = route

	baseExecutor := ocpbugs43745.NewPodExecutor(kubeClient, cfg)
	tc.PodExecutor = ocpbugs43745.NewLoggingPodExecutor(
		ocpbugs43745.NewRetryingExecutor(baseExecutor, 3, time.Second),
		logger,
	)

	return tc
}

func waitForHAProxyConfigCondition(tc *TestConfig, routeName string, service *corev1.Service, servicePod *corev1.Pod, shouldBePresent bool) error {
	pods, err := tc.KubeClient.CoreV1().Pods("openshift-ingress").List(tc.Context, metav1.ListOptions{
		LabelSelector: "ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default",
	})
	if err != nil {
		return fmt.Errorf("failed to list router pods: %w", err)
	}

	routerPods := make([]*routerPod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		routerPods = append(routerPods, &routerPod{
			name:      pod.Name,
			namespace: pod.Namespace,
			executor:  tc.PodExecutor,
		})
	}

	poller := newConfigPoller(5*time.Second, 2*time.Minute, tc.Logger)

	return poller.waitForCondition(tc.Context, func(ctx context.Context) (bool, error) {
		for _, pod := range routerPods {
			backends, err := pod.getHAProxyConfig(ctx)
			if err != nil {
				return false, err
			}

			backend, found := findBackend(backends, tc.Route, service, servicePod) // Adjust as needed
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

func waitForHAProxyUpdate(tc *TestConfig, routeName string, service *corev1.Service, pod *corev1.Pod) error {
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

	url := fmt.Sprintf("http://%s", route.Spec.Host)
	logger.Info("making GET request", "url", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
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

type routeUpdater struct {
	routeClient *routeclientset.Clientset
}

func (u *routeUpdater) Get(ctx context.Context, namespace, name string) (*routev1.Route, error) {
	return u.routeClient.RouteV1().Routes(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (u *routeUpdater) Update(ctx context.Context, route *routev1.Route) (*routev1.Route, error) {
	return u.routeClient.RouteV1().Routes(route.Namespace).Update(ctx, route, metav1.UpdateOptions{})
}

func updateRouteService(ctx context.Context, routeClient *routeclientset.Clientset, route *routev1.Route, newServiceName string, logger *slog.Logger) (*routev1.Route, error) {
	retryingUpdater := ocpbugs43745.NewRetryingUpdater(ocpbugs43745.NewLoggingUpdater(&routeUpdater{routeClient: routeClient}, logger))

	modifyFunc := func(r *routev1.Route) *routev1.Route {
		r.Spec.To = routev1.RouteTargetReference{
			Kind: "Service",
			Name: newServiceName,
		}
		return r
	}

	return retryingUpdater.Update(ctx, route, modifyFunc)
}

func switchServiceAndVerifyResponse(tc *TestConfig, service *corev1.Service, pod *corev1.Pod) error {
	updatedRoute, err := updateRouteService(tc.Context, tc.RouteClient, tc.Route, service.Name, tc.Logger)
	if err != nil {
		return err
	}

	if err := waitForHAProxyUpdate(tc, updatedRoute.Name, service, pod); err != nil {
		return err
	}

	return nil
}

func fetchServiceResponseAfterDelay(tc *TestConfig, delay time.Duration) (string, error) {
	time.Sleep(delay)
	return fetchServiceResponse(tc, tc.Route)
}

func TestServiceSwitcheroo(t *testing.T) {
	var delay time.Duration = 0

	if v := os.Getenv("SWITCH_DELAY"); v != "" {
		n, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("Invalid duration: %s: %v", v, err)
		}
		delay = n
	}

	tc := setupTest(t)
	defer func() {
		tc.cancel()
		if err := tc.Cleaner.Cleanup(); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}()

	t.Run("switching between services returns different responses", func(t *testing.T) {
		if len(tc.Services) < 2 || len(tc.Pods) < 2 {
			t.Fatal("Not enough services or pods to test switching")
		}

		// Switch to the first service.
		if err := switchServiceAndVerifyResponse(tc, tc.Services[0], &tc.Pods[0]); err != nil {
			t.Fatalf("Failed during switch to service %s: %v", tc.Services[0].Name, err)
		}
		tc.Logger.Info("Switched to service", "service", tc.Services[0].Name)

		resp1, err := fetchServiceResponseAfterDelay(tc, 0)
		if err != nil {
			t.Fatalf("Failed to fetch response from service %s: %v", tc.Services[0].Name, err)
		}

		// Switch to the second service.
		if err := switchServiceAndVerifyResponse(tc, tc.Services[1], &tc.Pods[1]); err != nil {
			t.Fatalf("Failed during switch to service %s: %v", tc.Services[1].Name, err)
		}
		tc.Logger.Info("Switched to service", "service", tc.Services[1].Name)

		resp2, err := fetchServiceResponseAfterDelay(tc, delay)
		if err != nil {
			t.Fatalf("Failed to fetch response from service %s: %v", tc.Services[1].Name, err)
		}

		if resp1 == resp2 {
			t.Fatalf("Expected different responses after switching services, but got the same response: %s", resp1)
		}
	})
}
