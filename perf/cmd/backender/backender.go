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

const b1024_html = `<!DOCTYPE html>
<html>
<head>
<title>Welcome 1024B</title>
</head>
<body>
<h1>Welcome 1024B</h1>

<pre>
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
1234567891123456789212345678931234567894123456789512345678961234567897123456789812345678991234567890123456789112345678921234567
</pre>

</body>
</html>
`

var (
	routes = flag.Int("routes", 5, "number of routes to standup")
)

type Backend struct {
	hostIPAddr string
	name       string
	port       int
}

// Program flags
var (
	hostPrefix = flag.String("host-prefxx", "http-scale", "prefix for hostname")
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

	for _, t := range perf.AllTerminationTypes {
		handler := http.FileServer(http.FS(htmlFS))

		for i := 0; i < *routes; i++ {
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
					if err := http.Serve(*l, handler); err != nil {
						log.Fatal(err)
					}
				default:
					if err := http.ServeTLS(*l, handler, tlsCertFile, tlsKeyFile); err != nil {
						log.Fatal(err)
					}
				}
			}(t, &ln)
		}
		printBackendConnectionInfo(os.Stdout, t)
	}

	backendsMux := http.NewServeMux()

	backendsMux.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
		for _, t := range perf.AllTerminationTypes[:] {
			printBackendConnectionInfo(w, t)
		}
	})

	backendsMux.HandleFunc("/backends/edge", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.EdgeTermination)
	})

	backendsMux.HandleFunc("/backends/http", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.HTTPTermination)
	})

	backendsMux.HandleFunc("/backends/passthrough", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.PassthroughTermination)
	})

	backendsMux.HandleFunc("/backends/reencrypt", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.ReEncryptTermination)
	})

	if err := http.ListenAndServe(":2000", backendsMux); err != nil {
		log.Fatal(err)
	}
}
