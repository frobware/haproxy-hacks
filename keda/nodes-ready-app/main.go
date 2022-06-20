package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"nodes-ready-app/pkg/autoscaler"

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

func init() {
	prometheus.DefaultRegisterer.MustRegister(readyNodesGauge)
	http.Handle("/metrics", promhttp.Handler())
}

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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("Terminating...")
				close(stopCh)
				return
			case <-time.After(3 * time.Second):
				var readyNodes float64
				var notReadyNodes float64
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
	}()

	wg.Wait()
}
