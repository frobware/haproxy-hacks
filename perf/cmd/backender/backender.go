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
	"strconv"

	"github.com/frobware/haproxy-hacks/perf"
)

type Backend struct {
	hostIPAddr string
	name       string
	port       int
}

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

	backends = map[perf.TerminationType][]Backend{}
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
	hostIPAddr, err := net.LookupIP(hostname)
	if err != nil {
		log.Fatal(err)
	}
	if len(hostIPAddr) == 0 {
		log.Fatalf("failed to resolve %q", hostname)
	}
	return fmt.Sprintf("%v", hostIPAddr[0])
}

func printBackendConnectionInfo(w io.Writer, t perf.TerminationType) error {
	for i := 0; i < len(backends[t]); i++ {
		io.WriteString(w, fmt.Sprintf("%v %v %v\n", backends[t][i].name, backends[t][i].hostIPAddr, backends[t][i].port))
	}
	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	hostIPAddr := mustResolveCurrentHost()

	tlsCertFile := tlsTemporaryCertFile()
	defer os.Remove(tlsCertFile)

	tlsKeyFile := tlsTemporaryKeyFile()
	defer os.Remove(tlsKeyFile)

	htmlHandler := http.FileServer(http.FS(htmlFS))

	for _, t := range perf.AllTerminationTypes {
		for i := 0; i < *nbackends; i++ {
			ln, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				log.Fatal(err)
			}
			backends[t] = append(backends[t], Backend{
				hostIPAddr: hostIPAddr,
				name:       fmt.Sprintf("%s-%v-%v", *hostPrefix, t, i),
				port:       ln.Addr().(*net.TCPAddr).Port,
			})
			go func(t perf.TerminationType, l *net.Listener) {
				switch t {
				case perf.HTTPTermination:
					if err := http.Serve(*l, htmlHandler); err != nil {
						log.Fatal(err)
					}
				default:
					if err := http.ServeTLS(*l, htmlHandler, tlsCertFile, tlsKeyFile); err != nil {
						log.Fatal(err)
					}
				}
			}(t, &ln)
		}
		printBackendConnectionInfo(os.Stdout, t)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
		for _, t := range perf.AllTerminationTypes[:] {
			printBackendConnectionInfo(w, t)
		}
	})
	mux.HandleFunc("/backends/edge", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.EdgeTermination)
	})
	mux.HandleFunc("/backends/http", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.HTTPTermination)
	})
	mux.HandleFunc("/backends/passthrough", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.PassthroughTermination)
	})
	mux.HandleFunc("/backends/reencrypt", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.ReEncryptTermination)
	})

	if err := http.ListenAndServe(":2000", mux); err != nil {
		log.Fatal(err)
	}
}
