package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"nodes-ready-app/pkg/autoscaler"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "", "path to the kubeconfig file")

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

type nodeLister struct {
	nodeInformer v1.NodeInformer
}

type NodeFilter func(*corev1.Node)

func NewNodeLister(client kubernetes.Interface, stopCh <-chan struct{}) (*nodeLister, error) {
	factory := informers.NewSharedInformerFactory(client, 0)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()

	go factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		return nil, errors.New("timed out waiting for caches to sync")
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return &nodeLister{nodeInformer: nodeInformer}, nil
}

func (l *nodeLister) ProcessNodes(filter NodeFilter) {
	nodes := l.nodeInformer.Informer().GetStore().List()

	for i := range nodes {
		if node, ok := (nodes[i].(*corev1.Node)); ok {
			filter(node)
		}
	}
}

func init() {
	prometheus.MustRegister(readyNodesGauge)
}

func main() {
	flag.Parse()
	if env := os.Getenv("KUBECONFIG"); env != "" {
		*kubeconfig = env
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan struct{})
	nodeLister, err := NewNodeLister(clientset, stopCh)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	http.Handle("/metrics", promhttp.Handler())

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(mainCtx)

	g.Go(func() error {
		switch err := server.ListenAndServe(); {
		case !errors.Is(err, http.ErrServerClosed):
			return fmt.Errorf("HTTP server error: %v", err)
		default:
			server.SetKeepAlivesEnabled(false)
			log.Println("Stopped serving new connections.")
			return nil
		}
	})

	g.Go(func() error {
		<-gCtx.Done()
		log.Println("WHOOOOOOOOOOOOOOOOOOOOOA Terminating...")
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), server.WriteTimeout)
		defer shutdownRelease()
		switch err := server.Shutdown(shutdownCtx); {
		case err == nil:
			log.Println("Graceful shutdown complete.")
			return nil
		default:
			return err
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				log.Println("Terminating...")
				close(stopCh)
				return nil
			case <-time.After(1 * time.Second):
				var readyNodes, notReadyNodes float64
				nodeLister.ProcessNodes(func(node *corev1.Node) {
					if autoscaler.IsNodeReadyAndSchedulable(node) {
						log.Println(node.GetName(), "READY")
						readyNodes += 1
					} else {
						log.Println(node.GetName(), "NOT-READY-OR-SCHEDULABLE")
						notReadyNodes += 1
					}
				})
				readyNodesGauge.Set(readyNodes)
				log.Println(readyNodes, "ready nodes")
				log.Println(notReadyNodes, "not ready nodes")
			}
		}
	})

	defer os.Exit(0)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("error: %s\n", err)
	}
}
