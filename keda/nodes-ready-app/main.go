package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"nodes-ready-app/pkg/autoscaler"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

func init() {
	prometheus.MustRegister(readyNodesGauge)
}

func processReadyNodes(store cache.Store) {
	var ready float64
	nodes := store.List()
	for i := range nodes {
		if node, ok := (nodes[i].(*corev1.Node)); ok {
			if autoscaler.IsNodeReadyAndSchedulable(node) {
				ready += 1
			}
		}
	}
	readyNodesGauge.Set(ready)
	klog.Infof("%v total nodes, %v ready nodes", len(nodes), ready)
}

func restConfig() (*rest.Config, error) {
	kubeCfg, err := rest.InClusterConfig()
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}
	return kubeCfg, nil
}

func main() {
	clusterConfig, err := restConfig()
	if err != nil {
		log.Fatalf("could not get config: %v\n", err)
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	http.Handle("/metrics", promhttp.Handler())

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{})
	go factory.Start(mainCtx.Done())
	if !cache.WaitForNamedCacheSync("controller", mainCtx.Done(), informer.HasSynced) {
		log.Fatalln("timed out waiting for caches to sync")
	}

	g, gCtx := errgroup.WithContext(mainCtx)

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		server.SetKeepAlivesEnabled(false)
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), server.WriteTimeout)
		defer shutdownRelease()
		return server.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil
			case <-time.After(2 * time.Second):
				processReadyNodes(informer.GetStore())
			}
		}
	})

	defer os.Exit(0)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}
