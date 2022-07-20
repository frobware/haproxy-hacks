package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/frobware/haproxy-hacks/keda/nodes-ready-app/pkg/autoscaler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/utils/net"
)

var updateIntervalFlag = flag.Duration("update-interval", time.Second,
	"interval between computing the number of ready nodes")

var versionFlag = flag.Bool("V", false, "Print program version and build info")

var readyNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ready_nodes",
	Help: "Report the number of Ready nodes in the cluster.",
})

func processReadyNodes(store cache.Store) (int, int) {
	var ready int
	var nodes = store.List()
	for i := range nodes {
		if node, ok := nodes[i].(*corev1.Node); ok {
			if autoscaler.IsNodeReadyAndSchedulable(node) {
				ready += 1
			}
		}
	}
	return len(nodes), ready
}

func restConfig() (*rest.Config, error) {
	kubeConfig, err := rest.InClusterConfig()
	if v := os.Getenv("KUBECONFIG"); v != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", v)
	}
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

func main() {
	prometheus.MustRegister(readyNodesGauge)
	klog.InitFlags(nil)
	flag.Parse()

	if *versionFlag {
		info, _ := debug.ReadBuildInfo()
		log.Println(info)
		os.Exit(0)
	}

	if val := os.Getenv("UPDATE_INTERVAL"); val != "" {
		if val, err := time.ParseDuration(val); err != nil {
			klog.Fatalf("failed to parse UPDATE_INTERVAL=%q: %v\n", val, err)
		} else {
			*updateIntervalFlag = val
		}
		klog.Infof("Setting UPDATE_INTERVAL=%q\n", *updateIntervalFlag)
	}

	port := "8080"
	if val := os.Getenv("PORT"); val != "" {
		if _, err := net.ParsePort(val, false); err != nil {
			klog.Fatalf("failed to parse PORT=%q: %v\n", val, err)
		}
		port = val
		klog.Infof("Setting PORT=%q\n", port)
	}

	clusterConfig, err := restConfig()
	if err != nil {
		log.Fatalf("could not get config: %v\n", err)
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := informerFactory.Core().V1().Nodes()
	sharedInformer := nodeInformer.Informer()
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, req *http.Request) {
		if !sharedInformer.HasSynced() {
			http.Error(w, "informers not synchronised", http.StatusInternalServerError)
			return
		}
		klog.Infoln(req)
		w.Write([]byte("/healthz/ready.\n"))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		klog.Infoln(req)
		w.Write([]byte("/healthz.\n"))
	})

	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%v", port),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	g, gCtx := errgroup.WithContext(signalCtx)

	g.Go(func() error {
		informerFactory.Start(gCtx.Done())
		if !cache.WaitForNamedCacheSync(path.Base(os.Args[0]), gCtx.Done(), sharedInformer.HasSynced) {
			return errors.New("timed out waiting for caches to sync")
		}
		return nil
	})

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
				if err := gCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil
			case <-time.After(*updateIntervalFlag):
				if sharedInformer.HasSynced() {
					total, ready := processReadyNodes(sharedInformer.GetStore())
					klog.Infof("%v nodes, %v ready, %v unready", total, ready, total-ready)
					readyNodesGauge.Set(float64(ready))
				}
			}
		}
	})

	defer klog.FlushAndExit(time.Second, 0)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.Flush()
		klog.Fatalln(err)
	}
}
