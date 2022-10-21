package main

import (
	"bytes"
	_ "embed"
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
	discoveryURL = flag.String("discovery", "http://localhost:2000", "backend discovery URL")
	httpPort     = flag.Int("http-port", 8080, "haproxy http port setting")
	httpsPort    = flag.Int("https-port", 8443, "haproxy https port setting")
	maxconn      = flag.Int("maxconn", 0, "haproxy maxconn setting")
	nbthread     = flag.Int("nbthread", 4, "haproxy nbthread setting")
	statsPort    = flag.Int("stats-port", 1936, "haproxy https port setting")
	configDir    = flag.String("config-dir", fmt.Sprintf("/tmp/%v-haproxy-gen", os.Getenv("USER")), "output path")
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

func filterBackendsByType(t perf.TerminationType, backends []BackendConfig) []BackendConfig {
	var result []BackendConfig

	for i := range backends {
		if backends[i].TerminationType == t {
			result = append(result, backends[i])
		}
	}

	return result
}

func writeFile(path string, data bytes.Buffer) error {
	dirname := filepath.Dir(path)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(data.Bytes())
	if err != nil {
		return err
	}
	return f.Close()
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

	if err := writeFile(path.Join(*configDir, "haproxy.config"), haproxyConf); err != nil {
		log.Fatal(err)
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
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_secure:%s\n", b.Name, b.Name)
			case perf.ReencryptTermination:
				return fmt.Sprintf("^%s\\.?(:[0-9]+)?(/.*)?$ be_edge_http:%s\n", b.Name, b.Name)
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

	for _, m := range maps {
		for _, t := range m.TerminationTypes {
			for _, b := range filterBackendsByType(t, allBackends) {
				if _, err := io.WriteString(m.Buffer, m.MapEntry(b)); err != nil {
					log.Fatal(err)
				}
			}
		}
		if err := writeFile(path.Join(config.ConfigDir, m.Filename), *m.Buffer); err != nil {
			log.Fatal(err)
		}
	}

	if err := writeFile(path.Join(*configDir, "error-page-404.http"), *bytes.NewBuffer([]byte(error404))); err != nil {
		log.Fatal(err)
	}

	if err := writeFile(path.Join(*configDir, "error-page-503.http"), *bytes.NewBuffer([]byte(error503))); err != nil {
		log.Fatal(err)
	}
}
