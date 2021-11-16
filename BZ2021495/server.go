package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("connection from %s\n", r.RemoteAddr)
		fmt.Fprintln(w, r.Proto, r.RemoteAddr)
	})

	for i := 9000; i < 9003; i++ {
		go func(port int) {
			addr := fmt.Sprintf(":%v", port)
			log.Printf("listening at %s\n", addr)
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				log.Fatalf("ListenAndServe: %v\n", err)
			}
		}(i)
	}

	select {}
}
