package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// handleConnection processes an incoming network connection, sending
// an HTTP response that either includes or excludes duplicate
// "Transfer-Encoding" headers, based on the duplicateTE flag. This
// function simulates an HTTP response, manually writing headers and
// chunked body data using a bufio.Writer.
//
// The reason we do not use http.ResponseWriter is that the Go HTTP
// stack automatically sanitises headers, preventing the inclusion of
// duplicate headers. For testing scenarios where we need to simulate
// multiple "Transfer-Encoding" headers, we manually write the
// response directly to the connection, giving us full control over
// the headers.
func handleConnection(conn net.Conn, duplicateTE bool) {
	defer conn.Close()

	log.Printf("%v: Connection from %s\n", conn.LocalAddr(), conn.RemoteAddr())

	writer := bufio.NewWriter(conn)

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Date: %s\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"Connection: close\r\n"+
		"Foo: Bar\r\n"+
		"Foo: Baz\r\n"+
		"Transfer-Encoding: chunked\r\n", time.Now().UTC().Format(time.RFC1123))

	if duplicateTE {
		response += "Transfer-Encoding: chunked\r\n"
	}

	response += "Foo: Baz\r\n" +
		"Foo: Bar\r\n" +
		"Set-Cookie: testcookie=value; path=/\r\n\r\n"

	if _, err := writer.WriteString(response); err != nil {
		log.Printf("%v: Error writing response headers: %v\n", conn.LocalAddr(), err)
		return
	}

	chunks := []string{
		"4\r\nTest\r\n",
		"A\r\nHelloWorld\r\n",
		"1\r\n\n\r\n",
		"0\r\n\r\n",
	}

	for i := range chunks {
		_, err := writer.WriteString(chunks[i])
		if err != nil {
			log.Printf("%v: Error writing chunk: %v\n", conn.LocalAddr(), err)
			break
		}
	}

	if err := writer.Flush(); err != nil {
		log.Printf("%v: Error flushing data to connection: %v\n", conn.LocalAddr(), err)
	}
}

func createListener(port string, useTLS bool) (net.Listener, error) {
	if useTLS {
		cert, err := tls.LoadX509KeyPair("/etc/serving-cert/tls.crt", "/etc/serving-cert/tls.key")
		if err != nil {
			return nil, fmt.Errorf("error loading TLS certificate: %w", err)
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		return tls.Listen("tcp", ":"+port, config)
	}

	return net.Listen("tcp", ":"+port)
}

func startServer(port string, servceDuplicateTransferEncodingHeader bool, useTLS bool) {
	listener, err := createListener(port, useTLS)
	if err != nil {
		log.Fatalf("Error starting server on port %s: %v\n", port, err)
	}

	defer listener.Close()

	fmt.Printf("Server is listening on port %v (useTLS=%v)\n", port, useTLS)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, servceDuplicateTransferEncodingHeader)
	}
}

func startHealthCheckServer(port string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	log.Printf("Health check server is listening on port %v\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting health check server: %v\n", err)
	}
}

func main() {
	singleTEPort := os.Getenv("SINGLE_TE_PORT")
	duplicateTEPort := os.Getenv("DUPLICATE_TE_PORT")
	singleTETLSPort := os.Getenv("SINGLE_TE_TLS_PORT")
	duplicateTETLSPort := os.Getenv("DUPLICATE_TE_TLS_PORT")
	healthCheckPort := os.Getenv("HEALTH_PORT")

	if singleTEPort == "" || duplicateTEPort == "" || singleTETLSPort == "" || duplicateTETLSPort == "" || healthCheckPort == "" {
		log.Fatalf("Environment variables SINGLE_TE_PORT, DUPLICATE_TE_PORT, SINGLE_TE_TLS_PORT, DUPLICATE_TE_TLS_PORT, and HEALTH_PORT must be set\n")
		os.Exit(1)
	}

	// Start non-TLS server with single Transfer-Encoding.
	go startServer(singleTEPort, false, false)

	// Start non-TLS server with duplicate Transfer-Encoding.
	go startServer(duplicateTEPort, true, false)

	// Start TLS server with single Transfer-Encoding.
	go startServer(singleTETLSPort, false, true)

	// Start TLS server with duplicate Transfer-Encoding.
	go startServer(duplicateTETLSPort, true, true)

	// Start the health check server.
	go startHealthCheckServer(healthCheckPort)

	// Block forever (needed to keep the servers running).
	select {}
}
