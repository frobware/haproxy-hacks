package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/frobware/g/nodes/pkg/autoscaler"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "", "path to the kubeconfig file")

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

const port = 9393

func init() {
	prometheus.DefaultRegisterer.MustRegister(readyNodesGauge)
}

func main() {
	flag.Parse()

	// override any CLI argument.
	if env := os.Getenv("KUBECONFIG"); env != "" {
		*kubeconfig = env
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	http.Handle("/metrics", promhttp.Handler())

	// stop signal for the informer
	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	defer runtime.HandleCrash()

	// start informer ->
	go factory.Start(stopper)

	// start to sync and call list
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: func(interface{}) { fmt.Println("delete not implemented") },
	})

	lister := nodeInformer.Lister()
	_, err = lister.List(labels.Everything())
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		for {
			var readyNodes float64
			nodes := informer.GetStore().List()
			for i := range nodes {
				n, ok := nodes[i].(*corev1.Node)
				if !ok {
					panic(fmt.Sprintf("%T", nodes[i]))
				}
				if autoscaler.IsNodeReadyAndSchedulable(n) {
					fmt.Println("READY", n.GetName())
					readyNodes += 1
				} else {
					if unreadiness, err := autoscaler.GetNodeReadiness(n); err != nil {
						fmt.Println("NOT READY", n.GetName(), unreadiness.Reason)
					}
				}
			}
			readyNodesGauge.Set(readyNodes)
			time.Sleep(2 * time.Second)
		}
	}()

	// fmt.Println("nodes:", nodes)

	log.Printf("Server started on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

	<-stopper

	// fmt.Println(clientset.CoreV1().Nodes().List(context.TODO(), corev1.ConfigMapList)
	//nodeInformer := kubeInformerFactory.Core().V1().Nodes().Informer()
}

func onAdd(obj interface{}) {
	if node, ok := obj.(*corev1.Node); ok {
		fmt.Println("ADD", node.GetName())
		fmt.Println(autoscaler.GetNodeReadiness(node))
	}
}

func onUpdate(prev, curr interface{}) {
	fmt.Printf("old %T, curr %T\n", prev, curr)
	prevNode, currNode := prev.(*corev1.Node), curr.(*corev1.Node)
	if diff := cmp.Diff(prevNode, currNode); diff != "" {
		fmt.Println("onUpdate", "\n", diff)
		fmt.Println(autoscaler.GetNodeReadiness(currNode))
	}
}
