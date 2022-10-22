package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/frobware/haproxy-hacks/perf"
)

type Backend struct {
	hostIPAddr      string
	name            string
	port            int
	terminationType perf.TerminationType
	listener        *net.Listener
}

type BackendsByTerminationType map[perf.TerminationType][]Backend

// Program flags
var (
	hostPrefix = flag.String("host-prefxx", "http-scale", "prefix for hostname")
	nbackends  = flag.Int("nbackends", 5, "number of backends servers to create")
)

var (
	//go:embed *.html
	htmlFS embed.FS

	//go:embed tls.crt
	tlsCert string

	//go:embed tls.key
	tlsKey string
)

func mustCreateTemporaryFile(data []byte) string {
	f, err := os.CreateTemp("", strconv.Itoa(os.Getpid()))
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return f.Name()
}

func tlsTemporaryKeyFile() string {
	return mustCreateTemporaryFile([]byte(tlsKey))
}

func tlsTemporaryCertFile() string {
	return mustCreateTemporaryFile([]byte(tlsCert))
}

func mustResolveCurrentHost() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}

func spawnBackend(b *Backend) {
	htmlHandler := http.FileServer(http.FS(htmlFS))

	go func() {
		tlsCertFile := tlsTemporaryCertFile()
		defer os.Remove(tlsCertFile)

		tlsKeyFile := tlsTemporaryKeyFile()
		defer os.Remove(tlsKeyFile)

		switch b.terminationType {
		case perf.HTTPTermination, perf.EdgeTermination:
			if err := http.Serve(*b.listener, htmlHandler); err != nil {
				log.Fatal(err)
			}
		default:
			if err := http.ServeTLS(*b.listener, htmlHandler, tlsCertFile, tlsKeyFile); err != nil {
				log.Fatal(err)
			}
		}
	}()
}

func startMetadataServer(port int) error {
	var backends = map[perf.TerminationType][]Backend{}

	printBackendsForType := func(w io.Writer, t perf.TerminationType) error {
		for i := 0; i < len(backends[t]); i++ {
			if _, err := io.WriteString(w, fmt.Sprintf("%v %v %v %v\n", backends[t][i].name, backends[t][i].hostIPAddr, backends[t][i].port, t)); err != nil {
				return err
			}
		}
		return nil
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
		for _, t := range perf.AllTerminationTypes[:] {
			printBackendsForType(w, t)
		}
	})
	mux.HandleFunc("/backends/edge", func(w http.ResponseWriter, r *http.Request) {
		printBackendsForType(w, perf.EdgeTermination)
	})
	mux.HandleFunc("/backends/http", func(w http.ResponseWriter, r *http.Request) {
		printBackendsForType(w, perf.HTTPTermination)
	})
	mux.HandleFunc("/backends/passthrough", func(w http.ResponseWriter, r *http.Request) {
		printBackendsForType(w, perf.PassthroughTermination)
	})
	mux.HandleFunc("/backends/reencrypt", func(w http.ResponseWriter, r *http.Request) {
		printBackendsForType(w, perf.ReencryptTermination)
	})

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), mux); err != nil {
		return err
	}

	return nil
}

func startBackends(backendTypes BackendsByTerminationType, pipeRdr, pipeWr *os.File) {
	var children = []int{}

	if err := startMetadataServer(2000); err != nil {
		log.Fatal(err)
	}

	log.SetPrefix(fmt.Sprintf("[P %v] ", os.Getpid()))
	log.Printf("pid: %d, ppid: %d, args: %s", os.Getpid(), os.Getppid(), os.Args)

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGCHLD)
		log.Println(<-sigc)
		os.Exit(1)
	}()

	var i = 0
	for t := range backendTypes {
		i += 1
		for _, backend := range backendTypes[t] {
			args := append(os.Args, fmt.Sprintf("#child_%d", i))
			childEnv := []string{fmt.Sprintf("BACKEND=%s", backend.name)}
			child, err := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
				Env:   append(os.Environ(), childEnv...),
				Files: []uintptr{0, 1, 2, pipeRdr.Fd()},
			})
			if err != nil {
				log.Fatal(err)
			}
			if child != 0 {
				children = append(children, child)
			}
		}
	}

	// for i := 0; i < fork; i++ {
	// 	args := append(os.Args, fmt.Sprintf("#child_%d_of_%d", i, fork))
	// 	childEnv := []string{
	// 		fmt.Sprintf("CHILD_ID=%d", i),
	// 	}
	// 	pwd, err := os.Getwd()
	// 	if err != nil {
	// 		log.Fatalf("getwd err: %s", err)
	// 	}
	// 	child, err := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
	// 		Dir:   pwd,
	// 		Env:   append(os.Environ(), childEnv...),
	// 		Files: []uintptr{0, 1, 2, pipeWr.Fd(), parentSentinel.Fd()},
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if child != 0 {
	// 		children = append(children, child)
	// 	}
	// }

	// log.Printf("children %+v", children)
	// buf := bufio.NewReader(pipeRd)

	// var remaining = len(children)

	// for remaining > 0 {
	// 	line, _, err := buf.ReadLine()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Println(remaining, string(line), err)
	// 	remaining -= 1
	// }

	log.Println("Waiting for READY!")

	select {}
}

func serveBackend(backend Backend, parentFD uintptr) {
	log.SetPrefix(fmt.Sprintf("[c %v] ", os.Getpid()))
	// n, err := io.WriteString(os.NewFile(parentFD, "<pipe>"), fmt.Sprintf("%v\n", os.Getpid()))
	// if err != nil {
	// 	log.Fatalf("fatal write: n=%v %v\n", n, err)
	// }
	n, err := os.NewFile(4, "<pipe>").Read(make([]byte, 1))
	log.Println("parent pipe wokeup", n, err)
	os.Exit(2)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	flag.Parse()

	hostIPAddr := mustResolveCurrentHost()
	backends := map[perf.TerminationType][]Backend{}

	for _, t := range perf.AllTerminationTypes {
		for i := 0; i < *nbackends; i++ {
			ln, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				log.Fatal(err)
			}
			backends[t] = append(backends[t], Backend{
				hostIPAddr:      hostIPAddr,
				name:            fmt.Sprintf("%s-%v-%v", *hostPrefix, t, i),
				port:            ln.Addr().(*net.TCPAddr).Port,
				terminationType: t,
				listener:        &ln,
			})
		}
	}

	if _, isChild := os.LookupEnv("CHILD_ID"); !isChild {
		pipeRd, pipeWr, err := os.Pipe()
		if err != nil {
			log.Fatal(err)
		}

		startBackends(backends, pipeRd, pipeWr)
		// for i := 0; i < fork; i++ {
		// 	args := append(os.Args, fmt.Sprintf("#child_%d_of_%d", i, fork))
		// 	childEnv := []string{
		// 		fmt.Sprintf("CHILD_ID=%d", i),
		// 	}
		// 	pwd, err := os.Getwd()
		// 	if err != nil {
		// 		log.Fatalf("getwd err: %s", err)
		// 	}
		// 	child, err := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
		// 		Dir:   pwd,
		// 		Env:   append(os.Environ(), childEnv...),
		// 		Files: []uintptr{0, 1, 2, pipeWr.Fd(), parentSentinel.Fd()},
		// 	})
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	if child != 0 {
		// 		children = append(children, child)
		// 	}
		// }

		// log.Printf("children %+v", children)
		// buf := bufio.NewReader(pipeRd)

		// var remaining = len(children)

		// for remaining > 0 {
		// 	line, _, err := buf.ReadLine()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	log.Println(remaining, string(line), err)
		// 	remaining -= 1
		// }

		// log.Println(remaining, "; all children are READY!")
	} else {
		// serveBackend(3)

		// log.SetPrefix(fmt.Sprintf("[c %v] ", os.Getpid()))
		// n, err := io.WriteString(os.NewFile(3, "<pipe>"), fmt.Sprintf("%v\n", os.Getpid()))
		// if err != nil {
		// 	log.Fatalf("fatal write: n=%v %v\n", n, err)
		// }
		// n, err = os.NewFile(4, "<pipe>").Read(make([]byte, 1))
		// log.Println("parent pipe wokeup", n, err)
		// os.Exit(2)
	}
}

func mainx() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	backends := map[perf.TerminationType][]Backend{}

	for _, t := range perf.AllTerminationTypes {
		for i := 0; i < *nbackends; i++ {
			ln, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				log.Fatal(err)
			}
			backends[t] = append(backends[t], Backend{
				//hostIPAddr:      hostIPAddr,
				name:            fmt.Sprintf("%s-%v-%v", *hostPrefix, t, i),
				port:            ln.Addr().(*net.TCPAddr).Port,
				terminationType: t,
				listener:        &ln,
			})
		}
	}

	//	hostIPAddr := mustResolveCurrentHost()

	// tlsCertFile := tlsTemporaryCertFile()
	// defer os.Remove(tlsCertFile)

	// tlsKeyFile := tlsTemporaryKeyFile()
	// defer os.Remove(tlsKeyFile)

	// for _, t := range perf.AllTerminationTypes {
	// 	for i := 0; i < *nbackends; i++ {
	// 		ln, err := net.Listen("tcp", "0.0.0.0:0")
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		backends[t] = append(backends[t], Backend{
	// 			hostIPAddr:      hostIPAddr,
	// 			name:            fmt.Sprintf("%s-%v-%v", *hostPrefix, t, i),
	// 			port:            ln.Addr().(*net.TCPAddr).Port,
	// 			terminationType: t,
	// 			listener:        &ln,
	// 		})
	// 		// go func(t perf.TerminationType, l *net.Listener) {
	// 		// 	switch t {
	// 		// 	case perf.HTTPTermination, perf.EdgeTermination:
	// 		// 		if err := http.Serve(*l, htmlHandler); err != nil {
	// 		// 			log.Fatal(err)
	// 		// 		}
	// 		// 	default:
	// 		// 		if err := http.ServeTLS(*l, htmlHandler, tlsCertFile, tlsKeyFile); err != nil {
	// 		// 			log.Fatal(err)
	// 		// 		}
	// 		// 	}
	// 		// }(t, &ln)
	// 	}
	// 	printBackendConnectionInfo(os.Stdout, t)
	// }

	// mux := http.NewServeMux()

	// mux.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
	// 	for _, t := range perf.AllTerminationTypes[:] {
	// 		printBackendConnectionInfo(w, t)
	// 	}
	// })
	// mux.HandleFunc("/backends/edge", func(w http.ResponseWriter, r *http.Request) {
	// 	printBackendConnectionInfo(w, perf.EdgeTermination)
	// })
	// mux.HandleFunc("/backends/http", func(w http.ResponseWriter, r *http.Request) {
	// 	printBackendConnectionInfo(w, perf.HTTPTermination)
	// })

	// mux.HandleFunc("/backends/passthrough", func(w http.ResponseWriter, r *http.Request) {
	// 	printBackendConnectionInfo(w, perf.PassthroughTermination)
	// })
	// mux.HandleFunc("/backends/reencrypt", func(w http.ResponseWriter, r *http.Request) {
	// 	printBackendConnectionInfo(w, perf.ReencryptTermination)
	// })

	// if err := http.ListenAndServe(":2000", mux); err != nil {
	// 	log.Fatal(err)
	// }
}
