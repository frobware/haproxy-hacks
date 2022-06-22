package main

import (
	"context"
	"errors"
	"flag"
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

var updateInterval = 15 * time.Second

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

func init() {
	prometheus.MustRegister(readyNodesGauge)
}

func processReadyNodes(store cache.Store) (int, int, int) {
	var ready int
	nodes := store.List()
	for i := range nodes {
		if node, ok := (nodes[i].(*corev1.Node)); ok {
			if autoscaler.IsNodeReadyAndSchedulable(node) {
				ready += 1
			}
		}
	}
	return len(nodes), ready, len(nodes) - ready
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
	klog.InitFlags(flag.CommandLine)
	klog.SetOutput(os.Stderr)

	clusterConfig, err := restConfig()
	if err != nil {
		log.Fatalf("could not get config: %v\n", err)
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}

	if val := os.Getenv("UPDATE_INTERVAL"); val != "" {
		if x, err := time.ParseDuration(val); err != nil {
			klog.Infof("failed to parse UPDATE_INTERVAL=%q: %v; defaulting to %v\n", val, err, updateInterval)
		} else {
			klog.Infof("Setting update interval to %s\n", x.String())
			updateInterval = x
		}
	}

	klog.Infof("update interval set to %v", updateInterval.String())

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	http.Handle("/metrics", promhttp.Handler())

	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := informerFactory.Core().V1().Nodes()
	sharedInformer := nodeInformer.Informer()
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	g, gCtx := errgroup.WithContext(mainCtx)

	go informerFactory.Start(gCtx.Done())
	if !cache.WaitForNamedCacheSync(os.Args[0], gCtx.Done(), sharedInformer.HasSynced) {
		klog.Fatalln("timed out waiting for caches to sync")
	}

	g.Go(func() error {
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		httpServer.SetKeepAlivesEnabled(false)
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), httpServer.WriteTimeout)
		defer shutdownRelease()
		return httpServer.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil
			case <-time.After(updateInterval):
				total, ready, unready := processReadyNodes(sharedInformer.GetStore())
				klog.Infof("%v total nodes, %v ready nodes, %v unready nodes", total, ready, unready)
				readyNodesGauge.Set(float64(ready))
			}
		}
	})

	defer os.Exit(0)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}
