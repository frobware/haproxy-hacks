package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/frobware/haproxy-hacks/perf"
	"github.com/frobware/haproxy-hacks/perf/certs"
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
	routes = flag.Int("routes", 100, "number of routes to standup")
)

type Backend struct {
	name     string
	port     int
	listener net.Listener
}

var (
	hostPrefix = flag.String("host-prefxx", "http-scale", "prefix for hostname")
)

var servers = map[perf.TerminationType][]Backend{}

func printBackendConnectionInfo(w io.Writer, t perf.TerminationType) error {
	for i := 0; i < len(servers[t]); i++ {
		io.WriteString(w, fmt.Sprintf("%v %v\n", servers[t][i].name, servers[t][i].port))
	}
	return nil
}

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	tlsCertFile := lookupEnv("TLS_CRT", certs.TLSCertFile())
	tlsKeyFile := lookupEnv("TLS_KEY", certs.TLSKeyFile())

	http.HandleFunc("/1024.html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(b1024_html))
	})

	http.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
		for _, t := range perf.AllTerminationTypes[:] {
			printBackendConnectionInfo(w, t)
		}
	})

	http.HandleFunc("/backends/edge", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.EdgeTermination)
	})

	http.HandleFunc("/backends/http", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.HTTPTermination)
	})

	http.HandleFunc("/backends/passthrough", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.PassthroughTermination)
	})

	http.HandleFunc("/backends/reencrypt", func(w http.ResponseWriter, r *http.Request) {
		printBackendConnectionInfo(w, perf.ReEncryptTermination)
	})

	for _, t := range perf.AllTerminationTypes {
		for i := 0; i < *routes; i++ {
			ln, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				log.Fatal(err)
			}
			backend := Backend{
				name:     fmt.Sprintf("%s-%v-%v", *hostPrefix, i, t),
				port:     ln.Addr().(*net.TCPAddr).Port,
				listener: ln,
			}
			servers[t] = append(servers[t], backend)
			go func(t perf.TerminationType, b *Backend) {
				switch t {
				case perf.HTTPTermination:
					if err := http.Serve(b.listener, nil); err != nil {
						log.Fatal(err)
					}
				default:
					if err := http.ServeTLS(b.listener, nil, tlsCertFile, tlsKeyFile); err != nil {
						log.Fatal(err)
					}
				}
			}(t, &backend)
		}
	}

	if err := http.ListenAndServe(":2000", nil); err != nil {
		log.Fatal(err)
	}
}
