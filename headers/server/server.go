/package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
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

		names := make([]string, 0, len(req.Header))
		for k := range req.Header {
			names = append(names, k)
		}

		sort.SliceStable(names, func(i, j int) bool {
			iTestHdr := strings.HasPrefix(names[i], "Test")
			jTestHdr := strings.HasPrefix(names[j], "Test")
			if iTestHdr && jTestHdr {
				ix, _ := strconv.Atoi(strings.Split(names[i], "_")[1])
				jx, _ := strconv.Atoi(strings.Split(names[j], "_")[1])
				return ix < jx
			}
			return names[i] < names[j]
		})

		for _, name := range names {
			log.Printf("%s: %q\n", name, req.Header[name])
		}

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
