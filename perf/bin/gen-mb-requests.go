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
	Name string
	Port int64
}

type Backends map[string]Backend

type TerminationType string

type RequestConfig struct {
	Clients           int64
	Domain            string
	KeepAliveRequests int64
	TLSSessionReuse   bool
	TerminationTypes  []TerminationType
}

var (
	tlsreuse = flag.Bool("tlsreuse", true, "enable TLS reuse")
	domain   = flag.String("domain", "", "domain name")
)

const (
	EdgeTermination        TerminationType = "edge"
	HTTPTermination        TerminationType = "http"
	PassthroughTermination TerminationType = "passthrough"
	ReEncryptTermination   TerminationType = "reencrypt"
)

var AllTerminationTypes = [...]TerminationType{
	EdgeTermination,
	HTTPTermination,
	PassthroughTermination,
	ReEncryptTermination,
}

func (t TerminationType) TerminationScheme() string {
	switch t {
	case HTTPTermination:
		return "http"
	default:
		return "https"
	}
}

func (t TerminationType) TerminationPort() int64 {
	switch t {
	case HTTPTermination:
		return 8080
	default:
		return 8443
	}
}

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
				Host:              fmt.Sprintf("%s-%s.%s", name, t, config.Domain),
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

	if *domain == "" {
		log.Fatal("no domain name specified")
	}

	scanner := bufio.NewScanner(os.Stdin)
	backends := Backends{}

	// Input format is lines of the form: <backend-name> <port-number>

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		port, err := strconv.ParseInt(words[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		backends[words[0]] = Backend{Name: words[0], Port: port}
	}

	for _, scenario := range []struct {
		Name             string
		TerminationTypes []TerminationType
	}{
		{"edge", []TerminationType{EdgeTermination}},
		{"http", []TerminationType{HTTPTermination}},
		{"mix", AllTerminationTypes[:]},
		{"passthrough", []TerminationType{PassthroughTermination}},
		{"reencrypt", []TerminationType{ReEncryptTermination}},
	} {
		for _, keepAliveRequests := range []int64{0, 1, 50} {
			config := RequestConfig{
				Clients:           100,
				Domain:            *domain,
				KeepAliveRequests: keepAliveRequests,
				TLSSessionReuse:   false,
				TerminationTypes:  scenario.TerminationTypes,
			}
			requests := generateRequests(config, backends)
			data, err := json.MarshalIndent(requests, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			filename := fmt.Sprintf("mb-backends:%v-clients:%v-keepalives:%v-%v.json", len(requests), config.Clients, config.KeepAliveRequests, scenario.Name)
			if err := writeFile(filename, data); err != nil {
				log.Fatalf("error generating %s: %v", filename, err)
			}
		}
	}
}
