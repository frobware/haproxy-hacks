package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func handleConnection(conn net.Conn, duplicateTE bool) {
	defer conn.Close()

	fmt.Printf("%v: Connection from %s\n", conn.LocalAddr(), conn.RemoteAddr())

	writer := bufio.NewWriter(conn)

	var response string
	if duplicateTE {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Date: %s\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"Connection: close\r\n"+ // Disable keep-alive
			"Foo: Bar\r\n"+
			"Foo: Baz\r\n"+
			"Transfer-Encoding: chunked\r\n"+
			"Transfer-Encoding: chunked\r\n"+ // Deliberate duplicate Transfer-Encoding header
			"Foo: Baz\r\n"+
			"Foo: Bar\r\n"+
			"Set-Cookie: testcookie=value; path=/\r\n"+
			"\r\n", time.Now().UTC().Format(time.RFC1123))
	} else {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Date: %s\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"Connection: close\r\n"+ // Disable keep-alive
			"Foo: Bar\r\n"+
			"Foo: Baz\r\n"+
			"Transfer-Encoding: chunked\r\n"+
			"Foo: Baz\r\n"+
			"Foo: Bar\r\n"+
			"Set-Cookie: testcookie=value; path=/\r\n"+
			"\r\n", time.Now().UTC().Format(time.RFC1123))
	}

	_, err := writer.WriteString(response)
	if err != nil {
		fmt.Printf("Error writing response headers: %v\n", err)
		return
	}

	chunks := []string{
		"4\r\nTest\r\n",
		"A\r\nHelloWorld\r\n",
		"1\r\n\n\r\n",
		"0\r\n\r\n", // Final chunk
	}

	for _, chunk := range chunks {
		_, err := writer.WriteString(chunk)
		if err != nil {
			fmt.Printf("Error writing chunk: %v\n", err)
			return
		}
	}

	// Make sure everything is written to the connection before
	// closing.
	if err := writer.Flush(); err != nil {
		fmt.Printf("%v: Error flushing data to connection: %v\n", conn.LocalAddr(), err)
		return
	}
}

func startServer(port string, duplicateTE bool, useTLS bool) {
	var listener net.Listener
	var err error

	if useTLS {
		cert, err := tls.LoadX509KeyPair("/etc/serving-cert/tls.crt", "/etc/serving-cert/tls.key")
		if err != nil {
			fmt.Printf("Error loading TLS certificate: %v\n", err)
			os.Exit(1)
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		listener, err = tls.Listen("tcp", ":"+port, config)
	} else {
		listener, err = net.Listen("tcp", ":"+port)
	}

	if err != nil {
		fmt.Printf("Error starting server on port %s: %v\n", port, err)
		return
	}

	defer listener.Close()

	fmt.Printf("Server is listening on :%s (useTLS=%v)\n", port, useTLS)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, duplicateTE)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func startHealthCheckServer(port string) {
	http.HandleFunc("/healthz", healthCheckHandler)
	fmt.Printf("Health check server is listening on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting health check server: %v\n", err)
	}
}

func main() {
	port1 := os.Getenv("SINGLE_TE_PORT")
	port2 := os.Getenv("DUPLICATE_TE_PORT")
	tlsPort1 := os.Getenv("SINGLE_TE_TLS_PORT")
	tlsPort2 := os.Getenv("DUPLICATE_TE_TLS_PORT")
	healthCheckPort := os.Getenv("HEALTH_PORT")

	if port1 == "" || port2 == "" || tlsPort1 == "" || tlsPort2 == "" || healthCheckPort == "" {
		fmt.Println("Environment variables SINGLE_TE_PORT, DUPLICATE_TE_PORT, SINGLE_TE_TLS_PORT, DUPLICATE_TE_TLS_PORT, and HEALTH_PORT must be set")
		os.Exit(1)
	}

	// Start the first server (normal Transfer-Encoding, no TLS).
	go startServer(port1, false, false)

	// Start the second server (duplicate Transfer-Encoding, no TLS).
	go startServer(port2, true, false)

	// Start the first TLS server (normal Transfer-Encoding, with TLS).
	go startServer(tlsPort1, false, true)

	// Start the second TLS server (duplicate Transfer-Encoding, with TLS).
	go startServer(tlsPort2, true, true)

	// Start the health check server.
	go startHealthCheckServer(healthCheckPort)

	select {}
}
