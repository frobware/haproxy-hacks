package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

var (
	discvoery = flag.String("backender", "http://localhost:2000", "backend discovery URL")
	httpPort  = flag.Int("http-port", 8080, "haproxy http port setting")
	httpsPort = flag.Int("https-port", 8443, "haproxy https port setting")
	maxconn   = flag.Int("maxconn", 0, "haproxy maxconn setting")
	nbthread  = flag.Int("nbthread", 4, "haproxy nbthread setting")
	statsPort = flag.Int("stats-port", 1936, "haproxy https port setting")
	configDir = flag.String("config-dir", fmt.Sprintf("/tmp/%s-haproxy-config", os.Getenv("USER")), "output path")
)

//go:embed preamble.tmpl
var preamble string

func main() {
	type HAProxySettings struct {
		ConfigDir string
		HTTPPort  int
		HTTPSPort int
		Maxconn   int
		Nbthread  int
		StatsPort int
	}
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	preambleTmpl, err := template.New("preamble").Parse(preamble)
	if err != nil {
		log.Fatal(err)
	}

	err = preambleTmpl.Execute(os.Stdout, HAProxySettings{
		ConfigDir: *configDir,
		HTTPPort:  *httpPort,
		HTTPSPort: *httpsPort,
		Maxconn:   *maxconn,
		Nbthread:  *nbthread,
		StatsPort: *statsPort,
	})

	if err != nil {
		log.Fatal(err)
	}
}
