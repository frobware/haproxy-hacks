package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/frobware/haproxy-hacks/perf"
)

type Request struct {
	Body              *Body    `json:"body,omitempty"`
	Clients           int64    `json:"clients"`
	Delay             Delay    `json:"delay"`
	Headers           *Headers `json:"headers,omitempty"`
	Host              string   `json:"host"`
	KeepAliveRequests int64    `json:"keep-alive-requests"`
	Method            string   `json:"method"`
	Path              string   `json:"path"`
	Port              int64    `json:"port"`
	Scheme            string   `json:"scheme"`
	TLSSessionReuse   bool     `json:"tls-session-reuse"`
}

type Body struct {
	Content string `json:"content"`
}

type Delay struct {
	Max int64 `json:"max"`
	Min int64 `json:"min"`
}

type Headers struct {
	ContentType string `json:"Content-Type"`
}

type Backend struct {
	Name     string
	HostAddr string
	Port     int64
}

type Backends map[string]Backend

type RequestConfig struct {
	Clients           int64
	KeepAliveRequests int64
	TLSSessionReuse   bool
	TerminationTypes  []perf.TerminationType
}

type HAProxyConfig struct {
	ConfigDir string
	HTTPPort  int
	HTTPSPort int
	Maxconn   int
	Nbthread  int
	StatsPort int
	Backends  []BackendConfig
}

type BackendConfig struct {
	BackendCookie   string
	ConfigDir       string
	HostAddr        string
	Name            string
	Port            string
	ServerCookie    string
	TerminationType perf.TerminationType
}

const (
	HTTPBackendMapName      = "os_http_be.map"
	ReencryptBackendMapName = "os_edge_reencrypt_be.map"
	SNIPassthroughMapName   = "os_sni_passthrough.map"
	TCPBackendMapName       = "os_tcp_be.map"
)

//go:embed globals.tmpl
var globalTemplate string

//go:embed defaults.tmpl
var defaultTemplate string

//go:embed backends.tmpl
var backendTemplate string

//go:embed error-page-404.http
var error404 string

//go:embed error-page-503.http
var error503 string

var (
	configDir    = flag.String("config-dir", fmt.Sprintf("/tmp/%v-haproxy-gen", os.Getenv("USER")), "output path")
	discoveryURL = flag.String("discovery", "http://localhost:2000", "backend discovery URL")
	httpPort     = flag.Int("http-port", 8080, "haproxy http port setting")
	httpsPort    = flag.Int("https-port", 8443, "haproxy https port setting")
	maxconn      = flag.Int("maxconn", 0, "haproxy maxconn setting")
	nbthread     = flag.Int("nbthread", 4, "haproxy nbthread setting")
	statsPort    = flag.Int("stats-port", 1936, "haproxy https port setting")
	tlsreuse     = flag.Bool("tlsreuse", true, "enable TLS reuse")
)

func cookie() string {
	letterRunes := []rune("0123456789abcdef")
	b := make([]rune, 32)
	for i := 0; i < 32; i++ {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func fetchBackendMetadata[T perf.TerminationType](t T) ([]string, error) {
	url := fmt.Sprintf("%s/backends/%v", *discoveryURL, t)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return strings.Split(strings.Trim(string(body), "\n"), "\n"), nil
	}

	return nil, fmt.Errorf("unexpected status %v", resp.StatusCode)
}

func fetchAllBackendMetadata() ([]BackendConfig, error) {
	var backends []BackendConfig

	for _, t := range perf.AllTerminationTypes {
		metadata, err := fetchBackendMetadata(t)
		if err != nil {
			return nil, err
		}
		for i := range metadata {
			words := strings.Split(metadata[i], " ")
			backends = append(backends, BackendConfig{
				BackendCookie:   cookie(),
				ConfigDir:       *configDir,
				HostAddr:        words[1],
				Name:            words[0],
				Port:            words[2],
				ServerCookie:    cookie(),
				TerminationType: t,
			})
		}
	}

	return backends, nil
}

func filterBackendsByType(types []perf.TerminationType, backends []BackendConfig) []BackendConfig {
	var result []BackendConfig

	for _, t := range types {
		for i := range backends {
			if backends[i].TerminationType == t {
				result = append(result, backends[i])
			}
		}
	}

	return result
}

func writeFile(path string, data []byte) error {
	dirname := filepath.Dir(path)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return f.Close()
}

func generateRequests(config RequestConfig, backends []BackendConfig) []Request {
	var requests []Request

	for _, b := range backends {
		requests = append(requests, Request{
			Clients:           config.Clients,
			Host:              fmt.Sprintf("%s", b.Name),
			KeepAliveRequests: config.KeepAliveRequests,
			Method:            "GET",
			Path:              "/1024.html",
			Port:              b.TerminationType.TerminationPort(),
			Scheme:            b.TerminationType.TerminationScheme(),
			TLSSessionReuse:   *tlsreuse,
		})
	}

	return requests
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if err := os.RemoveAll(*configDir); err != nil {
		log.Fatal(err)
	}

	allBackends, err := fetchAllBackendMetadata()
	if err != nil {
		log.Fatal(err)
	}

	config := HAProxyConfig{
		ConfigDir: *configDir,
		HTTPPort:  *httpPort,
		HTTPSPort: *httpsPort,
		Maxconn:   *maxconn,
		Nbthread:  *nbthread,
		StatsPort: *statsPort,
		Backends:  allBackends,
	}

	var haproxyConf bytes.Buffer

	for _, tmpl := range []*template.Template{
		template.Must(template.New("globals").Parse(globalTemplate)),
		template.Must(template.New("defaults").Parse(defaultTemplate)),
		template.Must(template.New("backends").Parse(backendTemplate)),
	} {
		if err := tmpl.Execute(&haproxyConf, config); err != nil {
			log.Fatal(err)
		}
	}

	type MapEntryFunc func(backend BackendConfig) string

	maps := []struct {
		Filename         string
		TerminationTypes []perf.TerminationType
		MapEntry         MapEntryFunc
		Buffer           *bytes.Buffer
	}{{
		Filename:         HTTPBackendMapName,
		TerminationTypes: []perf.TerminationType{perf.HTTPTermination},
		Buffer:           &bytes.Buffer{},
		MapEntry: func(b BackendConfig) string {
			switch b.TerminationType {
			case perf.HTTPTermination:
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_http:%s\n", b.Name, b.Name)
			default:
				panic(b.TerminationType)
			}
		},
	}, {
		Filename:         ReencryptBackendMapName,
		TerminationTypes: []perf.TerminationType{perf.ReencryptTermination, perf.EdgeTermination},
		Buffer:           &bytes.Buffer{},
		MapEntry: func(b BackendConfig) string {
			switch b.TerminationType {
			case perf.EdgeTermination:
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_edge_http:%s\n", b.Name, b.Name)
			case perf.ReencryptTermination:
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_secure:%s\n", b.Name, b.Name)
			default:
				panic(b.TerminationType)
			}
		},
	}, {
		Filename:         SNIPassthroughMapName,
		TerminationTypes: []perf.TerminationType{perf.PassthroughTermination},
		Buffer:           &bytes.Buffer{},
		MapEntry: func(b BackendConfig) string {
			switch b.TerminationType {
			case perf.PassthroughTermination:
				return fmt.Sprintf("^%s$ 1\n", b.Name)
			default:
				panic(b.TerminationType)
			}
		},
	}, {
		Filename:         TCPBackendMapName,
		TerminationTypes: []perf.TerminationType{perf.PassthroughTermination},
		Buffer:           &bytes.Buffer{},
		MapEntry: func(b BackendConfig) string {
			switch b.TerminationType {
			case perf.PassthroughTermination:
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_tcp:%s\n", b.Name, b.Name)
			default:
				panic(b.TerminationType)
			}
		},
	}}

	if err := writeFile(path.Join(*configDir, "conf", "haproxy.config"), haproxyConf.Bytes()); err != nil {
		log.Fatal(err)
	}

	if err := writeFile(path.Join(*configDir, "conf", "error-page-404.http"), bytes.NewBuffer([]byte(error404)).Bytes()); err != nil {
		log.Fatal(err)
	}

	if err := writeFile(path.Join(*configDir, "conf", "error-page-503.http"), bytes.NewBuffer([]byte(error503)).Bytes()); err != nil {
		log.Fatal(err)
	}

	for _, m := range maps {
		for _, b := range filterBackendsByType(m.TerminationTypes, allBackends) {
			if _, err := io.WriteString(m.Buffer, m.MapEntry(b)); err != nil {
				log.Fatal(err)
			}
		}
		if err := writeFile(path.Join(config.ConfigDir, "conf", m.Filename), m.Buffer.Bytes()); err != nil {
			log.Fatal(err)
		}
	}

	for _, clients := range []int64{1, 50, 100, 200} {
		for _, scenario := range []struct {
			Name             string
			TerminationTypes []perf.TerminationType
		}{
			{"edge", []perf.TerminationType{perf.EdgeTermination}},
			{"http", []perf.TerminationType{perf.HTTPTermination}},
			{"mix", perf.AllTerminationTypes[:]},
			{"passthrough", []perf.TerminationType{perf.PassthroughTermination}},
			{"reencrypt", []perf.TerminationType{perf.ReencryptTermination}},
		} {
			for _, keepAliveRequests := range []int64{0, 1, 50} {
				config := RequestConfig{
					Clients:           clients,
					KeepAliveRequests: keepAliveRequests,
					TLSSessionReuse:   false,
					TerminationTypes:  scenario.TerminationTypes,
				}
				requests := generateRequests(config, filterBackendsByType(scenario.TerminationTypes, allBackends))
				data, err := json.MarshalIndent(requests, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				path := fmt.Sprintf("%s/mb/traffic-%v-backends-%v-clients-%v-keepalives-%v",
					*configDir,
					scenario.Name,
					len(requests)/len(config.TerminationTypes),
					config.Clients,
					config.KeepAliveRequests)
				fmt.Println(path)
				if err := os.MkdirAll(path, 0755); err != nil {
					log.Fatalf("error: failed to create path: %q: %v", path, err)
				}
				filename := fmt.Sprintf("%s/requests.json", path)
				if err := writeFile(filename, data); err != nil {
					log.Fatalf("error generating %s: %v", filename, err)
				}
			}
		}
	}
}