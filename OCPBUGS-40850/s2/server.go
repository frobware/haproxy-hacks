// server.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

// NewServer creates a new custom server instance.
func NewServer() *Server {
	return &Server{
		mux: http.NewServeMux(),
	}
}

// Handle registers a handler for a specific route.
func (s *Server) Handle(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

// Serve starts listening on the provided address and manages
// connections.
func (s *Server) Serve(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles an individual connection, parsing requests and sending responses.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read and parse the request
	req, err := CustomRequestReader(conn)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		return
	}

	fmt.Println(req)

	// Use a custom response writer
	crw := NewCustomResponseWriter(conn, req)

	// Pass the request to the mux (router) for handling
	s.mux.ServeHTTP(crw, req)

	// Finalise response writing (e.g., closing connection or sending trailers)
	crw.Finalise(conn.LocalAddr())
}
