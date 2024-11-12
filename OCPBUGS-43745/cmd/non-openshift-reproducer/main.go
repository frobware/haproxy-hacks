/*
Package main implements a HAProxy backend switch tester. The program
tests HAProxy's behavior during frequent backend switching and reloads
to identify if HAProxy maintains stale connections to old backends
after reconfiguration and reload.

The program simulates backend switching by alternating between two
backend containers (`backend1` and `backend2`), each running an HTTP
server from an nginx-alpine image. HAProxy is configured to route
requests to one backend at a time, with switches happening in a loop.
After each backend switch and HAProxy reload, the program issues HTTP
requests to the HAProxy frontend to determine which backend served the
request.

If a request unexpectedly reaches the old backend, indicating that
HAProxy retained a stale connection, the program logs an error and
stops.

### Program Structure

1. **Initial Setup**:

   - The program starts two backend containers (`backend1` and
     `backend2`) running nginx-alpine HTTP servers.

   - HAProxy is configured initially to route requests to `backend1`.

   - The configuration and temporary files are managed in a temporary
     directory created at runtime.

2. **Backend Switching Logic**:

   - The program runs an infinite loop that alternates between
     `backend1` and `backend2` by modifying the HAProxy configuration
     file and reloading HAProxy with the updated settings.

   - After each switch, an HTTP request is sent to HAProxy on the
     frontend port, and the response is captured to identify which
     backend served the request.

3. **Verification and Delay**:

   - After each request, the program verifies that the response
     matches the expected backend.

   - The `-delay` flag specifies a delay between each backend switch
     to simulate real-world conditions.

   - The program retries verification up to a configurable timeout
     duration after each backend switch to ensure HAProxy responds
     with the expected backend.

4. **Cleanup**:

   - The program ensures all resources are cleaned up upon exit by
     stopping and removing the backend containers and deleting
     temporary files.

   - A signal handler listens for termination signals and triggers the
     cleanup process.

### Options

The following options are available:

- `-idle-close-on-response`: A boolean flag to enable or disable the
  `idle-close-on-response` option in HAProxy. When enabled, this can
  be used to test how HAProxy behaves with idle connections on
  response close. Defaults to true.

- `-delay`: Specifies the delay duration between each backend switch,
  allowing a simulation of real-world delays that may occur in
  production environments. Accepts standard Go duration formats (e.g.,
  `100ms`, `2s`). Defaults to no delay.

- `-check-haproxy`: A boolean flag to enable or disable checking if
  HAProxy is already running before starting the test. When enabled,
  prevents the test from starting if another HAProxy instance is
  detected. Defaults to true.
*/

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var sharedTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			DualStack: false,
			KeepAlive: 10 * time.Second,
		}
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		localAddr := conn.LocalAddr().String()
		remoteAddr := conn.RemoteAddr().String()
		log.Printf("Dial established: local=%s, remote=%s", localAddr, remoteAddr)

		return conn, nil
	},
	MaxIdleConns:        100,
	IdleConnTimeout:     90 * time.Second,
	DisableKeepAlives:   false,
	MaxIdleConnsPerHost: 10,
}

var sharedClient = &http.Client{
	Transport: sharedTransport,
	Timeout:   10 * time.Second,
}

// Backend represents a container running a backend service.
type Backend struct {
	Name  string
	Port  int
	ID    string
	Image string
}

// Config holds the test configuration.
type Config struct {
	Backends      []Backend
	FrontendPort  int
	HaproxyBin    string
	Podman        string
	RequestDelay  time.Duration
	TempDir       string
	PidFile       string
	ConfigFile    string
	IdleClose     bool
	cleanupCalled bool
}

// HAProxyManager handles HAProxy operations.
type HAProxyManager struct {
	config *Config
}

// ContainerManager handles container operations.
type ContainerManager struct {
	config *Config
}

// TestRunner orchestrates the test execution.
type TestRunner struct {
	config     *Config
	haproxy    *HAProxyManager
	containers *ContainerManager
}

func NewConfig(idleClose bool) *Config {
	tempDir, err := os.MkdirTemp("", "haproxy-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}

	nginxImage := "quay.io/openshifttest/nginx-alpine@sha256:04f316442d48ba60e3ea0b5a67eb89b0b667abf1c198a3d0056ca748736336a0"

	return &Config{
		Backends: []Backend{
			{Name: "backend1", Port: 18081, Image: nginxImage},
			{Name: "backend2", Port: 18082, Image: nginxImage},
		},
		FrontendPort:  18080,
		HaproxyBin:    os.Getenv("HAPROXY_BIN"),
		Podman:        "podman",
		RequestDelay:  0,
		TempDir:       tempDir,
		PidFile:       filepath.Join(tempDir, "haproxy.pid"),
		ConfigFile:    filepath.Join(tempDir, "haproxy.config"),
		IdleClose:     idleClose,
		cleanupCalled: false,
	}
}

func (c *Config) Cleanup() {
	if c.cleanupCalled {
		return
	}

	c.cleanupCalled = true

	configContent, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		log.Printf("Warning: could not read HAProxy config file %s: %v\n", c.ConfigFile, err)
	} else {
		log.Printf("HAProxy configuration at cleanup:\n%s\n", string(configContent))
	}

	for _, backend := range c.Backends {
		stopCmd := exec.Command(c.Podman, "stop", backend.Name)
		if err := stopCmd.Run(); err != nil {
			log.Printf("Warning: failed to stop container %s: %v\n", backend.Name, err)
		} else {
			log.Printf("Successfully stopped container %s\n", backend.Name)
		}

		rmCmd := exec.Command(c.Podman, "rm", backend.Name)
		if err := rmCmd.Run(); err != nil {
			log.Printf("Warning: failed to remove container %s: %v\n", backend.Name, err)
		} else {
			log.Printf("Successfully removed container %s\n", backend.Name)
		}
	}

	if pidBytes, err := os.ReadFile(c.PidFile); err == nil {
		pid := strings.TrimSpace(string(pidBytes))

		if killErr := exec.Command("kill", pid).Run(); killErr != nil {
			log.Printf("Warning: failed to kill process with PID %s: %v\n", pid, killErr)
		} else {
			log.Printf("Successfully killed process with PID %s\n", pid)
		}
	} else {
		log.Printf("Warning: could not read PID file %s: %v\n", c.PidFile, err)
	}

	os.RemoveAll(c.TempDir)
}

func NewContainerManager(config *Config) *ContainerManager {
	return &ContainerManager{config: config}
}

func (cm *ContainerManager) StartContainer(backend *Backend) error {
	cmd := exec.Command(cm.config.Podman, "ps", "-q", "-f", fmt.Sprintf("name=%s", backend.Name))
	if output, _ := cmd.Output(); len(output) > 0 {
		log.Printf("Container %s is already running.\n", backend.Name)

		return nil
	}

	cmd = exec.Command(cm.config.Podman, "ps", "-a", "-q", "-f", fmt.Sprintf("name=%s", backend.Name))
	if output, _ := cmd.Output(); len(output) > 0 {
		log.Printf("Container %s exists but is stopped. Starting it.\n", backend.Name)

		return exec.Command(cm.config.Podman, "start", backend.Name).Run()
	}

	log.Printf("Creating and starting container %s on port %d.\n", backend.Name, backend.Port)

	return exec.Command(cm.config.Podman, "run", "-d", "--name", backend.Name, "-p", fmt.Sprintf("%d:8080", backend.Port), backend.Image).Run()
}

func (cm *ContainerManager) GetContainerID(backend *Backend) (string, error) {
	cmd := exec.Command(cm.config.Podman, "ps", "-qf", fmt.Sprintf("name=%s", backend.Name))

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func NewHAProxyManager(config *Config) *HAProxyManager {
	return &HAProxyManager{config: config}
}

func (hm *HAProxyManager) WriteConfig(backend *Backend) error {
	defaults := `
global
    daemon
    log /dev/log local0 info

defaults
    mode http
    option httplog
    option dontlognull
    option log-health-checks
    timeout connect 5s
    timeout client 30s
    timeout server 50s
    timeout client-fin 1s
    timeout server-fin 1s
    timeout http-request 10s
    timeout http-keep-alive 300s`
	if hm.config.IdleClose {
		defaults += "\n    option idle-close-on-response"
	}

	config := fmt.Sprintf(`
%s

frontend fe_test
    log global
    bind *:%d
    default_backend be_service

backend be_service
    option httpchk GET /
    server %s 127.0.0.1:%d check inter 1s rise 2 fall 2
`, defaults, hm.config.FrontendPort, backend.Name, backend.Port)

	return os.WriteFile(hm.config.ConfigFile, []byte(config), 0644)
}

func (hm *HAProxyManager) Reload() error {
	pidBytes, err := os.ReadFile(hm.config.PidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w; ensure HAProxy is running and the PID file path is correctly specified", err)
	}

	oldPid := strings.TrimSpace(string(pidBytes))

	cmd := exec.Command(hm.config.HaproxyBin, "-f", hm.config.ConfigFile, "-p", hm.config.PidFile, "-sf", oldPid)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("HAProxy reload failed: %w\nOutput:\n%s", err, output.String())
	}

	log.Printf("HAProxy reload succeeded. Output:\n%s", output.String())

	go func() {
		for {
			if err := exec.Command("kill", "-0", oldPid).Run(); err != nil {
				log.Printf("Old HAProxy process (PID %s) has exited", oldPid)

				return
			}

			time.Sleep(1 * time.Millisecond)
		}
	}()

	return nil
}

func (hm *HAProxyManager) Start() error {
	cmd := exec.Command(hm.config.HaproxyBin, "-f", hm.config.ConfigFile, "-p", hm.config.PidFile)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start HAProxy: %w\nOutput:\n%s", err, output.String())
	}

	log.Printf("HAProxy started successfully. Output:\n%s", output.String())

	return nil
}

func NewTestRunner(config *Config) *TestRunner {
	return &TestRunner{
		config:     config,
		haproxy:    NewHAProxyManager(config),
		containers: NewContainerManager(config),
	}
}

func (tr *TestRunner) Initialize() error {
	origialDelay := tr.config.RequestDelay
	tr.config.RequestDelay = 0
	defer func() {
		tr.config.RequestDelay = origialDelay
	}()

	for i := range tr.config.Backends {
		backend := &tr.config.Backends[i]
		if err := tr.containers.StartContainer(backend); err != nil {
			return fmt.Errorf("failed to start %s: %w", backend.Name, err)
		}

		id, err := tr.containers.GetContainerID(backend)
		if err != nil {
			return fmt.Errorf("failed to get %s ID: %w", backend.Name, err)
		}

		backend.ID = id
		log.Printf("%s container ID: %s on port %d\n", backend.Name, backend.ID, backend.Port)
	}

	if err := tr.haproxy.WriteConfig(&tr.config.Backends[0]); err != nil {
		return fmt.Errorf("failed to write initial config: %w", err)
	}

	if err := tr.haproxy.Start(); err != nil {
		return fmt.Errorf("failed to start HAProxy: %w", err)
	}

	return tr.VerifyResponse(&tr.config.Backends[0], 0, 1)
}

func (tr *TestRunner) HitBackend() (string, error) {
	if tr.config.RequestDelay > 0 {
		log.Printf("Adding delay of %s before GET request", tr.config.RequestDelay)
		time.Sleep(tr.config.RequestDelay)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", tr.config.FrontendPort)

	response, err := fetchResponseWithTrace(url, slog.Default())
	if err != nil {
		return "", fmt.Errorf("failed to connect to frontend: %w", err)
	}

	return response, nil
}

func (tr *TestRunner) VerifyResponse(backend *Backend, retryInterval time.Duration, maxRetries int) error {
	return retryWithInterval(
		func() (bool, error) {
			response, err := tr.HitBackend()
			if err != nil {
				return false, fmt.Errorf("error in HitBackend: %w", err)
			}

			if strings.Contains(response, backend.ID) {
				log.Printf("Received expected response %q from backend %q", strings.TrimSpace(response), backend.Name)
				return true, nil
			}

			return false, fmt.Errorf("unexpected response %q; expected backend ID %q for backend %q", response, backend.ID, backend.Name)
		},
		maxRetries,
		retryInterval,
		fmt.Sprintf("Verify response for backend %q", backend.Name),
	)
}

func (tr *TestRunner) LoopBackends(retryInterval time.Duration, maxIterartions int) error {
	currentIndex := 1
	iteration := 0

	for {
		backend := &tr.config.Backends[currentIndex]
		log.Printf("Switching to backend %s on port %d\n", backend.Name, backend.Port)

		operationStart := time.Now()

		if err := tr.haproxy.WriteConfig(backend); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		writeTime := time.Since(operationStart)

		reloadStart := time.Now()
		if err := tr.haproxy.Reload(); err != nil {
			return fmt.Errorf("failed to reload HAProxy: %w", err)
		}
		reloadTime := time.Since(reloadStart)

		verifyStart := time.Now()
		maxRetries := 100
		if retryInterval == 0 {
			maxRetries = 1
		}
		if err := tr.VerifyResponse(backend, retryInterval, maxRetries); err != nil {
			return err
		}
		verifyTime := time.Since(verifyStart)

		totalTime := time.Since(operationStart)
		log.Printf("Operation timings for switch to %s:\n"+
			"  Config write:      %v\n"+
			"  HAProxy reload:    %v\n"+
			"  Backend verify:    %v\n"+
			"  Total switch time: %v\n",
			backend.Name,
			writeTime,
			reloadTime,
			verifyTime,
			totalTime)

		currentIndex = (currentIndex + 1) % len(tr.config.Backends)

		iteration += 1
		if iteration >= maxIterartions {
			return nil
		}
	}
}

func (tr *TestRunner) retryUntilTimeout(checkFunc func() error, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	attempt := 1

	for {
		attemptStart := time.Now()
		if err := checkFunc(); err == nil {
			log.Printf("Verification succeeded on attempt %d after %v",
				attempt, time.Since(attemptStart))
			return nil
		}
		log.Printf("Verification attempt %d took %v",
			attempt, time.Since(attemptStart))

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout reached after %v", timeout)
		}

		attempt++
		time.Sleep(1 * time.Millisecond)
	}
}

// retryWithInterval retries a provided function up to maxRetries
// times with a specified interval. It logs each attempt and returns
// either the result of a successful attempt or an error if max
// retries are reached.
func retryWithInterval(
	retryFunc func() (bool, error),
	maxRetries int,
	retryInterval time.Duration,
	description string,
) error {
	startTime := time.Now()
	for attempt := 1; attempt <= maxRetries; attempt++ {
		success, err := retryFunc()

		if success {
			log.Printf("%s succeeded on attempt %d after %v", description, attempt, time.Since(startTime))
			return nil
		}

		log.Printf("%s failed on attempt %d: %v", description, attempt, err)

		if attempt == maxRetries {
			return fmt.Errorf("%s failed after %d attempts and %v: %w", description, attempt, time.Since(startTime), err)
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("%s failed after %d retries", description, maxRetries)
}

// checkHAProxyRunning checks if any instance of the HAProxy binary is
// running. If it finds a running instance, it returns an error.
func checkHAProxyRunning(haproxyBin string) error {
	cmd := exec.Command("pgrep", "-f", haproxyBin)

	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return fmt.Errorf("found running instance(s) of %s:\n%s", haproxyBin, string(output))
	}

	return nil
}

// fetchResponseWithTrace performs a GET request to a specified URL
// with connection tracing and logs detailed connection and request
// information, using the shared client.
func fetchResponseWithTrace(url string, logger *slog.Logger) (string, error) {
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			logger.Info("Connection obtained",
				"reused", info.Reused,
				"wasIdle", info.WasIdle,
				"idleTime", info.IdleTime)
		},
		ConnectStart: func(network, addr string) {
			logger.Info("Starting connection", "network", network, "addr", addr)
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				logger.Error("Connection attempt failed", "network", network, "addr", addr, "error", err)
			} else {
				logger.Info("Connection successfully established", "network", network, "addr", addr)
			}
		},
		PutIdleConn: func(err error) {
			if err != nil {
				logger.Warn("Connection not returned to idle pool", "error", err)
			} else {
				logger.Info("Connection returned to idle pool")
			}
		},
		GotFirstResponseByte: func() {
			logger.Info("First byte of response received")
		},
	}

	ctx := httptrace.WithClientTrace(context.Background(), trace)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	logger.Info("Executing GET request", "url", url)

	resp, err := sharedClient.Do(req)
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

	logger.Info("Response", "status", resp.Status, "headers", resp.Header, "url", url, "body", string(body))

	return string(body), nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	haproxyBin := os.Getenv("HAPROXY_BIN")
	if haproxyBin == "" {
		log.Fatal("HAPROXY_BIN environment variable is not set (e.g., HAPROXY_BIN=haproxy)")
	}

	maxIterations := flag.Int("max-iterations", 1, "The maximum number of iterations for backend switching")
	checkHAProxy := flag.Bool("check-haproxy", true, "Check if HAProxy is already running before starting")
	retryInterval := flag.Duration("retry-interval", 0, "Duration to sleep on failed on responses")
	idleClose := flag.Bool("idle-close-on-response", true, "Enable/disable idle-close-on-response option in HAProxy")
	requestDelay := flag.Duration("delay", 0, "Duration to wait after switching backend and before sending request")

	flag.Parse()

	if *checkHAProxy {
		if err := checkHAProxyRunning(haproxyBin); err != nil {
			log.Fatal(err)
		}
	}

	config := NewConfig(*idleClose)
	config.RequestDelay = *requestDelay

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		config.Cleanup()
		os.Exit(1)
	}()

	runner := NewTestRunner(config)
	if err := runner.Initialize(); err != nil {
		log.Printf("Initialization failed: %v\n", err)

		return
	}

	if err := runner.LoopBackends(*retryInterval, *maxIterations); err != nil {
		log.Fatalf("LoopBackends failed: %v\n", err)
	}
}
