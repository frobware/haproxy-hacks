package main

import (
	"io"
	"log"
	"net/http"
)

// Handler for /single-te (always writes with chunked transfer encoding)
func handleSingleTE(w http.ResponseWriter, r *http.Request) {
	crw, ok := w.(*CustomResponseWriter)
	if ok {
		crw.EnableChunked()                            // Enable chunked encoding
		crw.AddTrailer("X-Single-Trailer", "SingleTE") // Add trailer
	}

	w.Write([]byte("This is a chunked response for /single-te"))
}

// Handler for /duplicate-te (adds duplicate Transfer-Encoding: chunked)
func handleDuplicateTE(w http.ResponseWriter, r *http.Request) {
	crw, ok := w.(*CustomResponseWriter)
	if ok {
		crw.EnableChunked() // Enable chunked encoding
		// Add duplicate Transfer-Encoding: chunked header
		crw.Header().Add("Transfer-Encoding", "chunked")
		crw.AddTrailer("X-Duplicate-Trailer", "DuplicateTE") // Add trailer
		crw.AddTrailer("Transfer-Encoding", "chunked")       // Add trailer
	}

	w.Write([]byte("This is a chunked response for /duplicate-te with duplicate Transfer-Encoding\n"))
}

func main() {
	server := NewServer()

	server.Handle("/single-te", http.HandlerFunc(handleSingleTE))
	server.Handle("/duplicate-te", http.HandlerFunc(handleDuplicateTE))

	server.Handle("/example", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if crw, ok := w.(*CustomResponseWriter); ok {
			crw.EnableChunked()
			crw.AddTrailer("X-Custom-Trailer", "TrailerValue")
		}

		w.Write([]byte("This is a chunked response"))
	}))

	server.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK\n"))
	}))

	server.Handle("/echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	if err := server.Serve(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
