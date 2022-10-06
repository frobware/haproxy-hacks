package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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

var (
	clients   = flag.Int("clients", 1, "number of clients")
	keepalive = flag.Int("keepalive", 0, "number of keepalive requests")
	scheme    = flag.String("scheme", "https", "either http or https")
	tlsreuse  = flag.Bool("tlsreuse", false, "enable TLS reuse")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var requests []Request
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		cfg := Request{
			Clients:           int64(*clients),
			Host:              scanner.Text(),
			KeepAliveRequests: int64(*keepalive),
			Method:            "GET",
			Path:              "/1024.html",
			Port:              8443,
			Scheme:            *scheme,
			TLSSessionReuse:   *tlsreuse,
		}
		requests = append(requests, cfg)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(requests, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
