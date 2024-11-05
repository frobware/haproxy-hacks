// Generic test helper machinery
package ocpbugs43745

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
)

type ResourceMeta struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

type Creator[T any] interface {
	Create(ctx context.Context, meta ResourceMeta) (*T, error)
}

type LoggingCreator[T any] struct {
	next   Creator[T]
	logger *slog.Logger
}

func NewLoggingCreator[T any](next Creator[T], logger *slog.Logger) Creator[T] {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingCreator[T]{next: next, logger: logger}
}

func (l *LoggingCreator[T]) Create(ctx context.Context, meta ResourceMeta) (*T, error) {
	l.logger.Info("creating resource", "type", fmt.Sprintf("%T", *new(T)), "namespace", meta.Namespace, "name", meta.Name)
	result, err := l.next.Create(ctx, meta)
	if err != nil {
		l.logger.Error("failed to create resource", "type", fmt.Sprintf("%T", *new(T)), "error", err)
	} else {
		l.logger.Info("successfully created resource", "type", fmt.Sprintf("%T", *new(T)))
	}
	return result, err
}

type ReadinessAwareCreator[T any] struct {
	next           Creator[T]
	readinessCheck ReadinessCheck[T]
	timeout        time.Duration
}

type ReadinessCheck[T any] func(context.Context, *T) (bool, error)

func NewReadinessAwareCreator[T any](
	next Creator[T],
	readinessCheck ReadinessCheck[T],
	timeout time.Duration,
) Creator[T] {
	return &ReadinessAwareCreator[T]{next: next, readinessCheck: readinessCheck, timeout: timeout}
}

func (r *ReadinessAwareCreator[T]) Create(ctx context.Context, meta ResourceMeta) (*T, error) {
	result, err := r.next.Create(ctx, meta)
	if err != nil {
		return nil, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return result, fmt.Errorf("resource not ready: %w", waitCtx.Err())
		case <-ticker.C:
			ready, err := r.readinessCheck(waitCtx, result)
			if err != nil {
				return result, fmt.Errorf("readiness check failed: %w", err)
			}
			if ready {
				return result, nil
			}
		}
	}
}

type DeepCopyable interface {
	metav1.Object
	runtime.Object
}

type Updater[T DeepCopyable] interface {
	Get(ctx context.Context, namespace, name string) (T, error)
	Update(ctx context.Context, obj T) (T, error)
}

type RetryingUpdater[T DeepCopyable] struct {
	updater Updater[T]
}

func NewRetryingUpdater[T DeepCopyable](updater Updater[T]) *RetryingUpdater[T] {
	return &RetryingUpdater[T]{updater: updater}
}

type ModifyFunc[T DeepCopyable] func(T) T

func (r *RetryingUpdater[T]) Update(ctx context.Context, resource T, modify ModifyFunc[T]) (T, error) {
	var updatedResource T
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latestResource, err := r.updater.Get(ctx, resource.GetNamespace(), resource.GetName())
		if err != nil {
			return fmt.Errorf("failed to get latest resource: %w", err)
		}
		resourceCopy, ok := latestResource.DeepCopyObject().(T)
		if !ok {
			return fmt.Errorf("failed to convert deep copy to type T")
		}
		updatedResource, err = r.updater.Update(ctx, modify(resourceCopy))
		return err
	})

	return updatedResource, err
}

type LoggingUpdater[T DeepCopyable] struct {
	next   Updater[T]
	logger *slog.Logger
}

func NewLoggingUpdater[T DeepCopyable](next Updater[T], logger *slog.Logger) Updater[T] {
	return &LoggingUpdater[T]{next: next, logger: logger}
}

func (l *LoggingUpdater[T]) Get(ctx context.Context, namespace, name string) (T, error) {
	l.logger.Info("getting resource", "namespace", namespace, "name", name)
	obj, err := l.next.Get(ctx, namespace, name)
	if err != nil {
		l.logger.Error("get operation failed", "error", err)
	} else {
		l.logger.Info("get operation successful", "resource", fmt.Sprintf("%T", obj), "namespace", obj.GetNamespace(), "name", obj.GetName())
	}
	return obj, err
}

func (l *LoggingUpdater[T]) Update(ctx context.Context, obj T) (T, error) {
	l.logger.Info("starting update operation", "resource", fmt.Sprintf("%T", obj), "namespace", obj.GetNamespace(), "name", obj.GetName())
	updatedObj, err := l.next.Update(ctx, obj)
	if err != nil {
		l.logger.Error("update operation failed", "error", err)
	} else {
		l.logger.Info("update operation successful", "resource", fmt.Sprintf("%T", updatedObj), "namespace", updatedObj.GetNamespace(), "name", updatedObj.GetName())
	}
	return updatedObj, err
}

func Int32Ptr(i int32) *int32 { return &i }

type CleanupFunc func() error

type Cleaner struct {
	funcs []CleanupFunc
	mu    sync.Mutex
}

func NewCleaner() *Cleaner {
	return &Cleaner{funcs: make([]CleanupFunc, 0)}
}

func (c *Cleaner) Add(f CleanupFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, f)
}

func (c *Cleaner) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for i := len(c.funcs) - 1; i >= 0; i-- {
		if err := c.funcs[i](); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

type PodExecutor interface {
	Execute(ctx context.Context, podName, namespace, container string, command []string) (string, string, error)
}

type k8sExecutor struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewPodExecutor(clientset *kubernetes.Clientset, config *rest.Config) PodExecutor {
	return &k8sExecutor{
		clientset: clientset,
		config:    config,
	}
}

func (e *k8sExecutor) Execute(ctx context.Context, podName, namespace, container string, command []string) (string, string, error) {
	req := e.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", container).
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false")

	for _, cmd := range command {
		req.Param("command", cmd)
	}

	exec, err := remotecommand.NewSPDYExecutor(e.config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to execute command: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}

type LoggingPodExecutor struct {
	next   PodExecutor
	logger *slog.Logger
}

func NewLoggingPodExecutor(next PodExecutor, logger *slog.Logger) PodExecutor {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingPodExecutor{next: next, logger: logger}
}

func (l *LoggingPodExecutor) Execute(ctx context.Context, podName, namespace, container string, command []string) (string, string, error) {
	l.logger.Info("executing pod command", "pod", podName, "namespace", namespace, "container", container, "command", command)
	stdout, stderr, err := l.next.Execute(ctx, podName, namespace, container, command)
	if err != nil {
		l.logger.Error("command failed", "pod", podName, "error", err)
	}
	return stdout, stderr, err
}

type RetryingExecutor struct {
	next     PodExecutor
	attempts int
	delay    time.Duration
}

func NewRetryingExecutor(next PodExecutor, attempts int, delay time.Duration) PodExecutor {
	return &RetryingExecutor{next: next, attempts: attempts, delay: delay}
}

func (r *RetryingExecutor) Execute(ctx context.Context, podName, namespace, container string, command []string) (stdout string, stderr string, err error) {
	for i := 0; i < r.attempts; i++ {
		stdout, stderr, err = r.next.Execute(ctx, podName, namespace, container, command)
		if err == nil {
			return stdout, stderr, nil
		}
		if i >= r.attempts-1 {
			return stdout, stderr, err
		}
		select {
		case <-ctx.Done():
			return stdout, stderr, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(r.delay):
		}
	}
	return stdout, stderr, err
}

type NamespaceCreator struct {
	ClientSet *kubernetes.Clientset
}

func (n *NamespaceCreator) Create(ctx context.Context, meta ResourceMeta) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: meta.Name,
			Labels:       meta.Labels,
		},
	}

	return n.ClientSet.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
}
