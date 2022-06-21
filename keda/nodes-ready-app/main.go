package main

import (
	"context"
	"errors"
	"fmt"
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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

type storeFilter func(*corev1.Node)

func init() {
	prometheus.MustRegister(readyNodesGauge)
}

func processNodes(store cache.Store) {
	var readyNodes, notReadyNodes float64
	nodes := store.List()
	for i := range nodes {
		node, ok := (nodes[i].(*corev1.Node))
		if !ok {
			panic(fmt.Sprintf("%T", nodes[i]))
			continue
		}
		if autoscaler.IsNodeReadyAndSchedulable(node) {
			log.Println(node.GetName(), "READY")
			readyNodes += 1
		} else {
			log.Println(node.GetName(), "NOT-READY-OR-SCHEDULABLE")
			notReadyNodes += 1
		}
	}
	readyNodesGauge.Set(readyNodes)
	log.Println(readyNodes, "ready nodes")
	log.Println(notReadyNodes, "not ready nodes")
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
	if !cache.WaitForCacheSync(mainCtx.Done(), informer.HasSynced) {
		log.Fatalln("timed out waiting for caches to sync")
	}

	g, gCtx := errgroup.WithContext(mainCtx)

	g.Go(func() error {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %v", err)
		}
		server.SetKeepAlivesEnabled(false)
		log.Println("Stopped serving new HTTP connections.")
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), server.WriteTimeout)
		defer shutdownRelease()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		log.Println("Graceful HTTP shutdown complete.")
		return nil
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil
			case <-time.After(2 * time.Second):
				processNodes(informer.GetStore())
			}
		}
	})

	defer os.Exit(0)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("error: %s\n", err)
	}
}
