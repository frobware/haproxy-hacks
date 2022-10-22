package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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

var (
	tlsreuse = flag.Bool("tlsreuse", true, "enable TLS reuse")
)

func writeFile(filename string, data []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Close()
}

func generateRequests(config RequestConfig, backends Backends) []Request {
	var requests []Request

	for _, t := range config.TerminationTypes {
		for name := range backends {
			requests = append(requests, Request{
				Clients:           config.Clients,
				Host:              fmt.Sprintf("%s", name),
				KeepAliveRequests: config.KeepAliveRequests,
				Method:            "GET",
				Path:              "/1024.html",
				Port:              t.TerminationPort(),
				Scheme:            t.TerminationScheme(),
				TLSSessionReuse:   *tlsreuse,
			})
		}
	}

	return requests
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	backends := Backends{}

	// Input format is lines of the form: <backend-name> <port-number>
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		port, err := strconv.ParseInt(words[2], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		backends[words[0]] = Backend{Name: words[0], HostAddr: words[1], Port: port}
	}

	for _, clients := range []int64{1, 4, 100, 200} {
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
				requests := generateRequests(config, backends)
				data, err := json.MarshalIndent(requests, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				path := fmt.Sprintf("mb/traffic/%v/backends/%v/clients/%v/keepalives/%v",
					scenario.Name,
					len(requests)/len(config.TerminationTypes),
					config.Clients,
					config.KeepAliveRequests)
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
