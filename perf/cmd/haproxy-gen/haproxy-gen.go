package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "embed"

	"github.com/frobware/haproxy-hacks/perf"
)

var (
	discovery = flag.String("backender", "http://localhost:2000", "backend discovery URL")
	httpPort  = flag.Int("http-port", 8080, "haproxy http port setting")
	httpsPort = flag.Int("https-port", 8443, "haproxy https port setting")
	maxconn   = flag.Int("maxconn", 0, "haproxy maxconn setting")
	nbthread  = flag.Int("nbthread", 4, "haproxy nbthread setting")
	statsPort = flag.Int("stats-port", 1936, "haproxy https port setting")
	configDir = flag.String("config-dir", fmt.Sprintf("/tmp/%v-haproxy-config", os.Getenv("USER")), "output path")
)

//go:embed preamble.tmpl
var preamble string

func getBackends[T perf.TerminationType](t T) ([]string, error) {
	url := fmt.Sprintf("%s/backends/%v", *discovery, t)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//defer resp.Body.Close()
	fmt.Println(resp)
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return strings.Split(string(body), "\n"), nil
	}

	return nil, fmt.Errorf("unexpected status %v", resp.StatusCode)
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

	fmt.Println(getBackends(perf.ReEncryptTermination))
}
