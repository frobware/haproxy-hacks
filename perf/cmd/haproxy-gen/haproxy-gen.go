package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/frobware/haproxy-hacks/perf"
)

type GlobalConfig struct {
	ConfigDir string
	HTTPPort  int
	HTTPSPort int
	Maxconn   int
	Nbthread  int
	StatsPort int
}

type Backends struct {
	Backends []BackendConfig
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

//go:embed preamble.tmpl
var preambleTemplate string

//go:embed backends.tmpl
var backendsTemplate string

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

func generateHTTPBackends(lines []string) {
	config := []BackendConfig{}

	for _, line := range lines {
		words := strings.Split(line, " ")
		_, err := strconv.ParseInt(words[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		config = append(config, BackendConfig{
			Name: words[0],
			Port: words[1],
		})
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// preambleTmpl, err := template.New("preamble").Parse(preamble)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = preambleTmpl.Execute(os.Stdout, Preamble{
	// 	ConfigDir: *configDir,
	// 	HTTPPort:  *httpPort,
	// 	HTTPSPort: *httpsPort,
	// 	Maxconn:   *maxconn,
	// 	Nbthread:  *nbthread,
	// 	StatsPort: *statsPort,
	// })

	// if err != nil {
	// 	log.Fatal(err)
	// }

	backendTmpl, err := template.New("backends").Parse(backendsTemplate)
	if err != nil {
		log.Fatal(err)
	}

	allBackendConfigs, err := fetchAllBackendMetadata()
	if err != nil {
		log.Fatal(err)
	}

	if err = backendTmpl.Execute(os.Stdout, Backends{
		Backends: allBackendConfigs,
	}); err != nil {
		log.Fatal(err)
	}
}
