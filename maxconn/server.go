package main

import (
	"io"
	"log"
	"net/http"
	"os/exec"
)

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	// log.Println(string(p))
	n, err = fw.w.Write(p)
	if err != nil {
		if _, ok := fw.w.(io.Closer); ok {
			log.Println(err)
			// f.Close()
			// panic("x")
		}
		log.Println(err)
		// return
	}
	if fw.f != nil {
		fw.f.Flush()
	}
	log.Println("sleeping...")
	// time.Sleep(time.Millisecond * 100)
	return
}

func main() {
	go func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			fw := flushWriter{w: w}
			if f, ok := w.(http.Flusher); ok {
				fw.f = f
			}
			cmd := exec.Command("find", "/home/aim/src/github.com/frobware/haproxy-hacks", "-print")
			cmd.Stdout = &fw
			cmd.Stderr = &fw
			log.Println(cmd.Run())
		}

		http.HandleFunc("/1", handler)
		server := &http.Server{
			Addr: ":4040",
			// ReadTimeout:  10 * time.Second,
			// WriteTimeout: 10 * time.Second,
		}
		server.SetKeepAlivesEnabled(false)
		log.Fatal(server.ListenAndServe())
	}()

	go func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			fw := flushWriter{w: w}
			if f, ok := w.(http.Flusher); ok {
				fw.f = f
			}
			cmd := exec.Command("find", "/home/aim/src/github.com/frobware/haproxy-hacks", "-print")
			cmd.Stdout = &fw
			cmd.Stderr = &fw
			log.Println(cmd.Run())
		}

		http.HandleFunc("/2", handler)
		server := &http.Server{
			Addr: ":4041",
			// ReadTimeout:  10 * time.Second,
			// WriteTimeout: 10 * time.Second,
		}
		server.SetKeepAlivesEnabled(false)
		log.Fatal(server.ListenAndServe())
	}()

	select {}
}
