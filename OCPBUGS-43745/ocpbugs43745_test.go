package ocpbugs43745_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var sharedTransport *http.Transport

type ocpbugs43745TestConfig struct {
	kubeClientset  *kubernetes.Clientset
	kubeConfig     *rest.Config
	logger         *slog.Logger
	routeClientset *routeclientset.Clientset

	namespace     string
	labels        map[string]string
	testRouteName string

	httpClientOptions routeClientOptions
	httpClient        *routeClient
}

// Resource getters that always fetch fresh resources.
type ResourceGetter struct {
	tc *ocpbugs43745TestConfig
}

func NewResourceGetter(tc *ocpbugs43745TestConfig) *ResourceGetter {
	return &ResourceGetter{tc: tc}
}

func (g *ResourceGetter) GetBackendPods(ctx context.Context) ([]*corev1.Pod, error) {
	podList, err := g.tc.kubeClientset.CoreV1().Pods(g.tc.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(g.tc.labels).String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	pods := make([]*corev1.Pod, 0, len(podList.Items))
	for i := range podList.Items {
		pods = append(pods, &podList.Items[i])
	}

	return pods, nil
}

func (g *ResourceGetter) GetServices(ctx context.Context) ([]*corev1.Service, error) {
	// Use only the app label to find all services.
	labelSelector := map[string]string{
		"app": "web-server",
	}

	serviceList, err := g.tc.kubeClientset.CoreV1().Services(g.tc.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelSelector).String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]*corev1.Service, 0, len(serviceList.Items))
	for i := range serviceList.Items {
		services = append(services, &serviceList.Items[i])
	}

	g.tc.logger.Info("Found services",
		"count", len(services),
		"namespace", g.tc.namespace)

	return services, nil
}

func (g *ResourceGetter) GetTestRoute(ctx context.Context) (*routev1.Route, error) {
	return g.tc.routeClientset.RouteV1().Routes(g.tc.namespace).Get(ctx, g.tc.testRouteName, metav1.GetOptions{})
}

// Helper to get specific service and its pod.
func (g *ResourceGetter) GetServiceAndPod(ctx context.Context, index int) (*corev1.Service, *corev1.Pod, error) {
	services, err := g.GetServices(ctx)
	if err != nil {
		return nil, nil, err
	}

	if index >= len(services) {
		return nil, nil, fmt.Errorf("service index %d out of bounds", index)
	}

	pods, err := g.tc.kubeClientset.CoreV1().Pods(g.tc.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(services[index].Spec.Selector).String(),
	})
	if err != nil {
		return nil, nil, err
	}

	if len(pods.Items) == 0 {
		return nil, nil, fmt.Errorf("no pods found for service %s", services[index].Name)
	}

	return services[index], &pods.Items[0], nil
}

type haproxyBackend struct {
	name     string
	settings []string
	servers  []string
}

// parseHAProxyConfig parses the HAProxy configuration content and
// returns a slice of haproxyBackend.
func parseHAProxyConfig(content string) ([]haproxyBackend, error) {
	var (
		backends       []haproxyBackend
		currentBackend *haproxyBackend
	)

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		if strings.HasPrefix(trimmedLine, "backend ") {
			if currentBackend != nil {
				backends = append(backends, *currentBackend)
			}

			name := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "backend"))
			if name == "" {
				return nil, fmt.Errorf("empty backend name on line %d", lineNum)
			}

			currentBackend = &haproxyBackend{
				name:     name,
				settings: []string{},
				servers:  []string{},
			}

			continue
		}

		if currentBackend == nil {
			continue
		}

		if strings.HasPrefix(trimmedLine, "server ") {
			currentBackend.servers = append(currentBackend.servers, trimmedLine)
		} else {
			currentBackend.settings = append(currentBackend.settings, trimmedLine)
		}
	}

	if currentBackend != nil {
		backends = append(backends, *currentBackend)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading HAProxy config: %w", err)
	}

	if len(backends) == 0 {
		return nil, errors.New("no backends found in configuration")
	}

	return backends, nil
}

// findBackend searches for a backend with the expected backend and
// server names. Returns the found backend and true if found, or an
// empty backend and false if not found.
func findBackend(backends []haproxyBackend, expectedBackendName, expectedServiceName string) (haproxyBackend, bool) {
	if expectedBackendName == "" || expectedServiceName == "" {
		return haproxyBackend{}, false
	}

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
	stdout, stderr, err := executeCommandInPod(ctx, p.kubeClient, p.restConfig, p.name, p.namespace, "router", []string{"cat", "/var/lib/haproxy/conf/haproxy.config"})
	if err != nil {
		return nil, fmt.Errorf("failed to get HAProxy config from pod %s: %w\nstderr: %s", p.name, err, stderr)
	}

	return parseHAProxyConfig(stdout)
}

type waitForAllRoutesAdmittedProgressFunc func(admittedRoutes, totalRoutes int, pendingRoutes []string)

// waitForAllRoutesAdmitted waits until all routes in the namespace
// are admitted.
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

// getRouterPods retrieves the router pods from the
// "openshift-ingress" namespace.
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

// waitForHAProxyConfigCondition waits until the HAProxy configuration
// meets the expected condition.
func waitForHAProxyConfigCondition(
	ctx context.Context,
	routerPods []*routerPod,
	expectedBackendName, expectedServerName string,
	shouldBePresent bool,
	logger *slog.Logger,
) error {
	return wait.PollUntilContextTimeout(ctx, 7*time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		for _, routerPod := range routerPods {
			backends, err := routerPod.getHAProxyConfig(ctx)
			if err != nil {
				return false, err
			}

			backend, found := findBackend(backends, expectedBackendName, expectedServerName)

			if found == shouldBePresent {
				if found {
					logger.Info("HAProxy backend entry FOUND",
						"pod", routerPod.name,
						"backend", expectedBackendName,
						"servers", strings.Join(backend.servers, " "))
				} else {
					logger.Info("Backend entry absent as expected",
						"pod", routerPod.name,
						"route", expectedBackendName)
				}
			} else {
				logger.Info("HAProxy backend entry NOT found",
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

// waitForHAProxyConfigUpdate waits for the HAProxy configuration to
// update after switching services.
func waitForHAProxyConfigUpdate(
	ctx context.Context,
	kubeClient *kubernetes.Clientset,
	restConfig *rest.Config,
	route *routev1.Route,
	service *corev1.Service,
	backendPod *corev1.Pod,
	logger *slog.Logger,
) error {
	logger.Info("Waiting for HAProxy configuration update",
		"service", service.Name,
		"backend", fmt.Sprintf("be_http:%s:%s", route.Namespace, route.Name),
		"server", fmt.Sprintf("pod:%s:%s", backendPod.Name, service.Name))

	routerPods, err := getRouterPods(kubeClient, restConfig)
	if err != nil {
		return fmt.Errorf("failed to get router pods: %w", err)
	}

	expectedBackendName := fmt.Sprintf("be_http:%s:%s", route.Namespace, route.Name)
	expectedServerName := fmt.Sprintf("pod:%s:%s", backendPod.Name, service.Name)

	err = waitForHAProxyConfigCondition(ctx, routerPods, expectedBackendName, expectedServerName, true, logger)
	if err != nil {
		return fmt.Errorf("failed waiting for HAProxy configuration update: %w", err)
	}

	return nil
}

func fetchServiceResponse(logger *slog.Logger, route *routev1.Route, client *routeClient) (string, error) {
	response, err := client.getResponse(route)
	if err != nil {
		logger.Error("Failed getting response from service",
			"service", route.Spec.To.Name,
			"host", route.Spec.Host,
			"error", err)

		return "", fmt.Errorf("failed to get response from service: %w", err)
	}

	logger.Info("Received response from service",
		"service", route.Spec.To.Name,
		"host", route.Spec.Host,
		"response", response,
		"namespace", route.Namespace,
		"routeName", route.Name)

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

	if err := waitForRouteAdmission(ctx, routeClient, updatedRoute.Namespace, updatedRoute.Name); err != nil {
		return nil, fmt.Errorf("route not admitted after service switch: %w", err)
	}

	logger.Info("Route service switch complete",
		"route", updatedRoute.Name,
		"service", service.Name)

	return updatedRoute, nil
}

func waitForRouteAdmission(ctx context.Context, routeClient *routeclientset.Clientset, namespace, name string) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		route, err := routeClient.RouteV1().Routes(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		admitted := false
		ready := false

		for _, ingress := range route.Status.Ingress {
			if ingress.RouterCanonicalHostname != "" {
				admitted = true

				for _, condition := range ingress.Conditions {
					if condition.Type == routev1.RouteAdmitted && condition.Status == corev1.ConditionTrue {
						ready = true

						break
					}
				}
			}
		}

		return admitted && ready, nil
	})
}

// switchRouteServiceAndFetchResponse switches the route to a
// specified service index and fetches the response.
func switchRouteServiceAndFetchResponse(
	ctx context.Context,
	tc *ocpbugs43745TestConfig,
	serviceIndex int,
	delay time.Duration,
) (string, error) {
	getter := NewResourceGetter(tc)

	service, backendPod, err := getter.GetServiceAndPod(ctx, serviceIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get service and pod: %w", err)
	}

	route, err := getter.GetTestRoute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get test route: %w", err)
	}

	updatedRoute, err := routeSwitchServiceAndVerifyResponse(
		ctx,
		tc.kubeClientset,
		tc.kubeConfig,
		tc.routeClientset,
		route,
		service,
		backendPod,
		tc.logger,
	)
	if err != nil {
		return "", fmt.Errorf("failed during switch to service %s: %w", service.Name, err)
	}

	tc.testRouteName = updatedRoute.Name

	if err := waitForRouteAdmission(ctx, tc.routeClientset, updatedRoute.Namespace, updatedRoute.Name); err != nil {
		return "", fmt.Errorf("route admission failed: %w", err)
	}

	if delay > 0 {
		tc.logger.Info("Delaying GET request", "route", tc.testRouteName, "duration", delay)
		time.Sleep(delay)
	}

	currentRoute, err := getter.GetTestRoute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current route: %w", err)
	}

	return fetchServiceResponse(tc.logger, currentRoute, tc.httpClient)
}

// waitForReplicationControllerReady waits for the replication
// controller to have the desired number of ready replicas.
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

type routeClientOptions struct {
	EnableKeepAlive bool
	CacheControl    bool
	Timeout         time.Duration
}

type routeClient struct {
	client  *http.Client
	logger  *slog.Logger
	options routeClientOptions
}

type routeResponse struct {
	statusCode int
	body       string
	headers    http.Header
}

func newRouteClient(options routeClientOptions, logger *slog.Logger) *routeClient {
	if sharedTransport == nil {
		sharedTransport = &http.Transport{
			DisableKeepAlives: !options.EnableKeepAlive,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}

				conn, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}

				localAddr := conn.LocalAddr().String()
				remoteAddr := conn.RemoteAddr().String()
				logger.Info("Connection", "local", localAddr, "remote", remoteAddr)

				return conn, nil
			},
		}
	}

	return &routeClient{
		client: &http.Client{
			Timeout:   options.Timeout,
			Transport: sharedTransport,
		},
		logger:  logger,
		options: options,
	}
}

func (c *routeClient) getResponse(route *routev1.Route) (string, error) {
	if err := c.validateRoute(route); err != nil {
		return "", err
	}

	url := c.buildURL(route)

	response, err := c.executeRequest(route, url)
	if err != nil {
		return "", err
	}

	return response.body, nil
}

func (c *routeClient) validateRoute(route *routev1.Route) error {
	if route == nil {
		return errors.New("route cannot be nil")
	}

	if route.Spec.Host == "" {
		return fmt.Errorf("route %s/%s has no host", route.Namespace, route.Name)
	}

	return nil
}

func (c *routeClient) buildURL(route *routev1.Route) string {
	return "http://" + route.Spec.Host
}

func (c *routeClient) executeRequest(route *routev1.Route, url string) (*routeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.options.Timeout)
	defer cancel()

	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			c.logger.Info("Connection info",
				"reused", info.Reused,
				"wasIdle", info.WasIdle,
				"idleTime", info.IdleTime)
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				c.logger.Error("Connection failed",
					"network", network,
					"addr", addr,
					"error", err)
			} else {
				c.logger.Info("Connection established",
					"network", network,
					"addr", addr)
			}
		},
		PutIdleConn: func(err error) {
			if err != nil {
				c.logger.Warn("Connection not put back to idle",
					"error", err)
			} else {
				c.logger.Info("Connection returned to idle pool")
			}
		},
	}

	ctx = httptrace.WithClientTrace(ctx, trace)

	req, err := c.createRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.options.CacheControl {
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Pragma", "no-cache")
	}

	c.logger.Info("Making HTTP request",
		"url", url,
		"service", route.Spec.To.Name,
		"keepAliveDisabled", sharedTransport.DisableKeepAlives,
		"cacheControl", c.options.CacheControl)

	// Execute the request.
	resp, err := c.doRequest(req)
	if err != nil {
		c.logger.Error("Request failed",
			"url", url,
			"error", err,
			"routeName", route.Name,
			"service", route.Spec.To.Name)

		return nil, err
	}
	defer resp.Body.Close()

	// Log connection reuse info if available.
	if resp.Header.Get("X-Connection-Info") != "" {
		c.logger.Info("Connection info",
			"info", resp.Header.Get("X-Connection-Info"),
			"url", url)
	}

	// Process and return the response.
	return c.processResponse(resp)
}

func (c *routeClient) createRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *routeClient) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	return resp, nil
}

func (c *routeClient) processResponse(resp *http.Response) (*routeResponse, error) {
	if err := c.validateStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	body, err := c.readBody(resp.Body)
	if err != nil {
		return nil, err
	}

	c.logger.Info("Received response",
		"status", resp.Status,
		"headers", fmt.Sprintf("%+v", resp.Header))

	return &routeResponse{
		statusCode: resp.StatusCode,
		body:       body,
		headers:    resp.Header,
	}, nil
}

func (c *routeClient) validateStatusCode(statusCode int) error {
	if statusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	return nil
}

func (c *routeClient) readBody(body io.ReadCloser) (string, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}

func setupKubernetesClients(_ context.Context, tc *ocpbugs43745TestConfig) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	kubeClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	routeClientset, err := routeclientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create route client: %w", err)
	}

	tc.kubeConfig = cfg
	tc.kubeClientset = kubeClientset
	tc.routeClientset = routeClientset

	return nil
}

func logClusterVersion(ctx context.Context, tc *ocpbugs43745TestConfig) error {
	configClientset, err := configclientset.NewForConfig(tc.kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create config clientset: %w", err)
	}

	clusterVersion, err := configClientset.ConfigV1().ClusterVersions().Get(ctx, "version", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster version: %w", err)
	}

	tc.logger.Info("Running test on OpenShift Cluster Version",
		"version", clusterVersion.Status.Desired.Version)

	return nil
}

func setupNamespace(ctx context.Context, tc *ocpbugs43745TestConfig) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "route-service-switcher-test-",
			Labels:       tc.labels,
		},
	}

	created, err := tc.kubeClientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	tc.namespace = created.Name

	return nil
}

func setupBackendServices(ctx context.Context, tc *ocpbugs43745TestConfig) error {
	for i := 1; i <= 2; i++ {
		if err := createBackendService(ctx, tc, i); err != nil {
			return fmt.Errorf("failed to create backend %d: %w", i, err)
		}
	}

	return nil
}

func createReplicationController(ctx context.Context, tc *ocpbugs43745TestConfig, index int, labels map[string]string) (*corev1.ReplicationController, error) {
	int32Ptr := func(i int32) *int32 { return &i }

	rc := &corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("web-server-%d", index),
			Namespace: tc.namespace,
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

	created, err := tc.kubeClientset.CoreV1().ReplicationControllers(tc.namespace).Create(ctx, rc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create ReplicationController %d: %w", index, err)
	}

	tc.logger.Info("Created ReplicationController",
		"name", created.Name,
		"namespace", created.Namespace,
		"labels", labels)

	return created, nil
}

func createService(ctx context.Context, tc *ocpbugs43745TestConfig, index int, labels map[string]string) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("service-%d", index),
			Namespace: tc.namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(8080),
				},
			},
			Selector: labels,
		},
	}

	created, err := tc.kubeClientset.CoreV1().Services(tc.namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Service %d: %w", index, err)
	}

	tc.logger.Info("Created Service",
		"name", created.Name,
		"namespace", created.Namespace,
		"labels", labels)

	return created, nil
}

func createBackendService(ctx context.Context, tc *ocpbugs43745TestConfig, index int) error {
	instanceLabel := strconv.Itoa(index)
	labels := map[string]string{
		"app":      "web-server",
		"instance": instanceLabel,
	}

	rc, err := createReplicationController(ctx, tc, index, labels)
	if err != nil {
		return err
	}

	if err := waitForReplicationControllerReady(ctx, tc.kubeClientset, rc, 2*time.Minute); err != nil {
		return fmt.Errorf("RC %d is not ready: %w", index, err)
	}

	svc, err := createService(ctx, tc, index, labels)
	if err != nil {
		return err
	}

	_, err = createRoute(ctx, tc.routeClientset, tc.namespace, fmt.Sprintf("svc-%d", index), svc.Name, labels)
	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	return nil
}

func setupTestRoute(ctx context.Context, tc *ocpbugs43745TestConfig) error {
	getter := NewResourceGetter(tc)

	services, err := getter.GetServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

	if len(services) == 0 {
		return errors.New("no services found")
	}

	route, err := createRoute(ctx, tc.routeClientset, tc.namespace, "test", services[0].Name, tc.labels)
	if err != nil {
		return fmt.Errorf("failed to create test route: %w", err)
	}

	tc.testRouteName = route.Name

	return nil
}

func getEnvBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}

	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}

	return defaultVal
}

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

	tc := &ocpbugs43745TestConfig{
		logger: slog.Default(),
		labels: map[string]string{
			"test": "route-service-switch",
			"app":  "web-server",
		},
		httpClientOptions: routeClientOptions{
			EnableKeepAlive: getEnvBool("ENABLE_KEEPALIVE", true),
			CacheControl:    getEnvBool("USE_CACHE_CONTROL", true),
			Timeout:         getEnvDuration("CLIENT_TIMEOUT", 10*time.Second),
		},
	}

	tc.httpClient = newRouteClient(tc.httpClientOptions, tc.logger)

	if err := setupKubernetesClients(ctx, tc); err != nil {
		t.Fatalf("failed to setup clients: %v", err)
	}

	if err := logClusterVersion(ctx, tc); err != nil {
		t.Fatalf("failed to log cluster version: %v", err)
	}

	if err := setupNamespace(ctx, tc); err != nil {
		t.Fatalf("failed to setup namespace: %v", err)
	}

	if err := setupBackendServices(ctx, tc); err != nil {
		t.Fatalf("failed to setup backend services: %v", err)
	}

	if err := setupTestRoute(ctx, tc); err != nil {
		t.Fatalf("failed to setup test route: %v", err)
	}

	if v := os.Getenv("NO_CLEANUP"); v != "1" {
		t.Cleanup(func() {
			if !t.Failed() {
				tc.kubeClientset.CoreV1().Namespaces().Delete(context.Background(), tc.namespace, metav1.DeleteOptions{})
			} else {
				t.Logf("Test failed; leaving test setup in place in namespace %s", tc.namespace)
			}
		})
	}

	t.Run("switching between services returns different responses", func(t *testing.T) {
		tc.logger.Info("Testing with HTTP client options:",
			"EnableKeepAlive", tc.httpClientOptions.EnableKeepAlive,
			"CacheControl", tc.httpClientOptions.CacheControl,
			"Timeout", tc.httpClientOptions.Timeout)

		resp1, err := switchRouteServiceAndFetchResponse(ctx, tc, 0, 0)
		if err != nil {
			t.Fatal(err)
		}

		resp2, err := switchRouteServiceAndFetchResponse(ctx, tc, 1, getResponseDelay)
		if err != nil {
			t.Fatal(err)
		}

		if resp1 == resp2 {
			t.Errorf("Expected different responses after switching services, but got the same response: %s", resp1)

			// Keep trying until we get a different response or timeout
			logger := tc.logger.With("phase", "retry")

			retryCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()

			logger.Info("Starting retry loop to wait for service switch to take effect")

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			getter := NewResourceGetter(tc)
			attempts := 0

			for {
				select {
				case <-retryCtx.Done():
					t.Fatalf("Timed out waiting for service switch. All responses matched original: %s", resp1)

					return
				case <-ticker.C:
					attempts++
					logger.Info("Retrying request", "attempt", attempts)

					route, err := getter.GetTestRoute(retryCtx)
					if err != nil {
						logger.Error("Failed to get route during retry", "error", err)

						continue
					}

					newResp, err := fetchServiceResponse(logger, route, tc.httpClient)
					if err != nil {
						logger.Error("Failed to get response during retry",
							"error", err,
							"route", route.Name,
							"service", route.Spec.To.Name)

						continue
					}

					if newResp != resp1 {
						logger.Info("Successfully got different response",
							"originalResponse", resp1,
							"newResponse", newResp,
							"attempts", attempts,
							"route", route.Name,
							"service", route.Spec.To.Name)

						return
					}

					logger.Info("Still getting original response",
						"response", newResp,
						"attempts", attempts,
						"route", route.Name,
						"service", route.Spec.To.Name)
				}
			}
		}
	})
}
