package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

const (
	defaultHTTPPort = "3264"
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
	connectionCh := make(chan bool)
	doTicker := lookupEnv("TICKER", "false") == "true"

	if doTicker {
		ticker := time.Tick(1 * time.Second)
		go func() {
			var connections int64
			for {
				select {
				case <-connectionCh:
					connections += 1
				case <-ticker:
					log.Printf("connection/s: %v", connections)
					connections = 0
				}
			}

		}()
	}

	busyTime, err := time.ParseDuration(lookupEnv("BUSY_TIMEOUT", "0s"))

	if err != nil {
		log.Fatalf("failed to parse BUSY_TIMEOUT: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleConnStart := time.Now()
		host, port, _ := net.SplitHostPort(r.RemoteAddr)
		atomic.AddInt64(&clientCon, 1)
		n := clientCon
		if doTicker {
			connectionCh <- true
		}
		log.Printf("connection %v host %v port %v\n", n, host, port)

		readAllStart := time.Now()
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		readAllDone := time.Now()

		rs := RequestSummary{
			URL:     r.URL.RequestURI(),
			Method:  r.Method,
			Headers: r.Header,
			Params:  r.URL.Query(),
			Auth:    r.URL.User,
			Body:    string(bytes),
		}

		resp, err := json.MarshalIndent(&rs, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if busyTime != 0 {
			time.Sleep(busyTime)
		}

		writeStart := time.Now()
		w.Write(resp)
		w.Write([]byte("\n"))
		writeDone := time.Now()

		queryid := "0"
		if val, ok := rs.Params["queryid"]; ok {
			if len(val) > 0 {
				queryid = val[0]
			}
		}

		token := "0"
		if val, ok := rs.Params["token"]; ok {
			if len(val) > 0 {
				token = val[0]
			}
		}

		log.Printf("c-complete %v host %v port %v queryid %v busytime %f readbody %.9f writeresp %.9f total %.9f token %v\n",
			n,
			host,
			port,
			queryid,
			busyTime.Seconds(),
			readAllDone.Sub(readAllStart).Seconds(),
			writeDone.Sub(writeStart).Seconds(),
			time.Now().Sub(handleConnStart).Seconds(),
			token)
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
