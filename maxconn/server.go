package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	fmt.Println("sleeping...")
	time.Sleep(time.Millisecond * 250)
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	cmd := exec.Command("find", "/home/aim/src/github.com/frobware/haproxy-hacks", "-print")
	cmd.Stdout = &fw
	cmd.Stderr = &fw
	fmt.Println(cmd.Run())
}

func main() {
	http.HandleFunc("/", handler)
	server := &http.Server{Addr: ":4040"}
	server.SetKeepAlivesEnabled(false)
	log.Fatal(server.ListenAndServe())
}