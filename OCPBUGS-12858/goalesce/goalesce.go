package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"os"
)

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
			ForceAttemptHTTP2: true, // Attempt to use HTTP/2, but fall back to HTTP/1.1 if needed
		},
	}

	var connInfo *httptrace.GotConnInfo

	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			connInfo = &info
		},
	}

	if len(os.Args) == 0 {
		os.Exit(1)
	}

	for _, url := range os.Args[1:] {
		for i := 0; i < 10; i++ {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				panic(err)
			}
			req.URL.Scheme = "https"
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error fetching %s: %v\n", url, err)
				continue
			}

			proto := "HTTP/1.1"
			if resp.ProtoMajor == 2 {
				proto = "HTTP/2"
			}
			fmt.Printf("Fetched %s, protocol: %s, HTTP status: %d, reused connection: %v\n", url, proto, resp.StatusCode, connInfo.Reused)
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}
