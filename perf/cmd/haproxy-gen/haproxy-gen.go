package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/frobware/haproxy-hacks/perf"
)

type HTTPBackend struct {
	Name         string
	Port         int
	Cookie       string
	ServerCookie string
}

type BackendConfig struct {
	Type         string
	Name         string
	Port         string
	Cookie       string
	ServerCookie string
}

//go:embed preamble.tmpl
var preambleTemplate string

//go:embed backends.tmpl
var backendsTemplate string

var (
	discovery = flag.String("backender", "http://localhost:2000", "backend discovery URL")
	httpPort  = flag.Int("http-port", 8080, "haproxy http port setting")
	httpsPort = flag.Int("https-port", 8443, "haproxy https port setting")
	maxconn   = flag.Int("maxconn", 0, "haproxy maxconn setting")
	nbthread  = flag.Int("nbthread", 4, "haproxy nbthread setting")
	statsPort = flag.Int("stats-port", 1936, "haproxy https port setting")
	configDir = flag.String("config-dir", fmt.Sprintf("/tmp/%v-haproxy-config", os.Getenv("USER")), "output path")
)

func backendMetadata[T perf.TerminationType](t T) ([]string, error) {
	url := fmt.Sprintf("%s/backends/%v", *discovery, t)
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
		return strings.Split(string(body), "\n"), nil
	}

	return nil, fmt.Errorf("unexpected status %v", resp.StatusCode)
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
	type Preamble struct {
		ConfigDir string
		HTTPPort  int
		HTTPSPort int
		Maxconn   int
		Nbthread  int
		StatsPort int
	}
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

	_, err := template.New("backends").Parse(backendsTemplate)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(backendMetadata(perf.ReEncryptTermination))
}
