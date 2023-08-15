package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	defaultHealthPort = "4242"
	defaultHTTPPort   = "8080"
)

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func writeResponse(w http.ResponseWriter, req *http.Request) {
	if false {
		fmt.Printf("Headers: ")
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Printf("%v: %v\n", name, h)
			}
		}
		fmt.Println()
	}
	fmt.Fprintf(w, "Source port: %s | %s %s %s://%s:%s%s\n",
		strings.Split(req.RemoteAddr, ":")[1],
		req.Proto,
		req.RemoteAddr,
		req.Header.Get("X-Forwarded-Proto"),
		req.Header.Get("X-Forwarded-Host"),
		req.Header.Get("X-Forwarded-Port"),
		req.URL)
}

func main() {
	http.HandleFunc("/healthy", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	http.HandleFunc("/api", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	log.Println("GODEBUG", lookupEnv("GODEBUG", "<unset>"))

	go func() {
		port := lookupEnv("HTTP_PORT", defaultHTTPPort)
		log.Printf("Listening on port %v\n", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		port := lookupEnv("HEALTH_PORT", defaultHealthPort)
		log.Printf("Listening on port %v for health checks\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}
