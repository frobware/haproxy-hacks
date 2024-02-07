package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var host string
	var count int
	var delaySec int
	var newClient bool
	var insecure bool
	var scheme string
	var cycle bool

	flag.StringVar(&host, "host", "", "Host to GET")
	flag.IntVar(&count, "count", 30, "Number of GET requests to perform")
	flag.IntVar(&delaySec, "delay", 1, "Delay between requests in seconds")
	flag.BoolVar(&newClient, "newclient", false, "Create a new client for each request to avoid keepalives")
	flag.BoolVar(&insecure, "k", false, "Allow insecure server connections when using SSL")
	flag.StringVar(&scheme, "scheme", "http", "URL scheme (http or https)")
	flag.BoolVar(&cycle, "cycle", false, "Cycle the resolved host address before each GET request")
	flag.Parse()

	if host == "" {
		fmt.Println("Host is required")
		os.Exit(1)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}

	if len(ips) == 0 {
		panic("No IPs found for host")
	}

	responseDistribution := make(map[string]int)

	j := 0

	for i := 0; i < count; i++ {
		if cycle {
			j = (j + 1) % len(ips)
		}
		url := fmt.Sprintf("%s://%s", scheme, ips[j])
		client := createClient(newClient, insecure)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}

		req.Host = host
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stderr, "Request #%d to %s (%s) completed with status code: %d\n", i+1, host, ips[j], resp.StatusCode)

		bodyStr := string(body)
		bodyStr = strings.ReplaceAll(bodyStr, "\n", "")
		bodyStr = strings.ReplaceAll(bodyStr, "\r", "")
		fmt.Println(bodyStr)

		responseDistribution[bodyStr]++

		time.Sleep(time.Duration(delaySec) * time.Second)
	}

	fmt.Println("Distribution of responses:")
	for response, count := range responseDistribution {
		fmt.Printf("Response: %q - Count: %d\n", response, count)
	}
}

func createClient(newClient, insecure bool) *http.Client {
	transport := &http.Transport{
		DisableKeepAlives: newClient,
	}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{Transport: transport}
}
