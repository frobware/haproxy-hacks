package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	defaultHTTPPort  = "1936"
	defaultHTTPSPort = "8443"
	defaultTLSCrt    = "/etc/serving-cert/tls.crt"
	defaultTLSKey    = "/etc/serving-cert/tls.key"
)

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	crtFile := lookupEnv("TLS_CRT", defaultTLSCrt)
	keyFile := lookupEnv("TLS_KEY", defaultTLSKey)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Println("connection from", req.RemoteAddr)
		w.Header().Set("set-cookie2", "X=Y")
		fmt.Fprint(w, req.Proto)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		log.Println("connection from", req.RemoteAddr)
		fmt.Fprint(w, "healthz")
		fmt.Fprint(w, req.Proto, "\n")
	})

	go func() {
		port := lookupEnv("HTTP_PORT", defaultHTTPPort)
		log.Printf("Listening on port %v\n", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		port := lookupEnv("HTTPS_PORT", defaultHTTPSPort)
		log.Printf("Listening securely on port %v\n", port)

		if err := http.ListenAndServeTLS(":"+port, crtFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}
