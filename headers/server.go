package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	defaultHTTPPort = "9090"
)

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Println("connection from", req.RemoteAddr)
		for k, v := range req.Header {
			log.Printf("%s: %v\n", k, v)
		}
		w.Header().Set("set-cookie2", "X=Y")
		fmt.Fprintln(w, req.Proto)
		log.Println()
		log.Println()
	})

	port := lookupEnv("HTTP_PORT", defaultHTTPPort)
	log.Printf("Listening on port %v\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
