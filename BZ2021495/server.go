package main

import (
	"fmt"
	"log"
	"net/http"
)

type handler struct {
	Port int

	http.Handler
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%d connection from %s\n", h.Port, r.RemoteAddr)
	fmt.Fprintln(w, r.Proto, r.RemoteAddr)
}

func main() {
	for i := 9000; i < 9006; i++ {
		go func(port int) {
			addr := fmt.Sprintf(":%v", port)
			log.Printf("listening at %s\n", addr)
			err := http.ListenAndServe(addr, handler{Port: port})
			if err != nil {
				log.Fatalf("ListenAndServe: %v\n", err)
			}
		}(i)
	}

	select {}
}
