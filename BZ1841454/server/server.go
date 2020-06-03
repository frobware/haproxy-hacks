package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

const (
	defaultHTTPPort = "8080"
)

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

type RequestSummary struct {
	URL     string
	Method  string
	Headers http.Header
	Params  url.Values
	Auth    *url.Userinfo
	Body    string
}

var clientCon int64 = 0
var randomSrc = rand.NewSource(time.Now().Unix())

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	busyTime, err := time.ParseDuration(lookupEnv("BUSY_TIMEOUT", "0s"))

	if err != nil {
		log.Fatalf("failed to parse BUSY_TIMEOUT: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		atomic.AddInt64(&clientCon, 1)
		n := clientCon
		log.Println("connection", n, r.RemoteAddr)
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rs := RequestSummary{
			URL:     r.URL.RequestURI(),
			Method:  r.Method,
			Headers: r.Header,
			Params:  r.URL.Query(),
			Auth:    r.URL.User,
			// Body:    string(bytes),
		}

		resp, err := json.MarshalIndent(&rs, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if busyTime != 0 {
			time.Sleep(busyTime)
		}

		w.Write(resp)
		w.Write([]byte("\n"))

		log.Println("c-complete", n, r.RemoteAddr, busyTime, time.Now().Sub(now))
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "ready")
	})

	port := lookupEnv("HTTP_PORT", defaultHTTPPort)
	log.Printf("Listening on port %v\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
