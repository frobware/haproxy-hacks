package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func handleConnection(conn net.Conn, duplicateTE bool) {
	defer conn.Close()

	var response string
	if duplicateTE {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Date: %s\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"Foo: Bar\r\n"+
			"Transfer-Encoding: chunked\r\n"+
			"Transfer-Encoding: chunked\r\n"+ // deliberate Duplicate Transfer-Encoding header
			"Foo: Baz\r\n"+
			"Set-Cookie: testcookie=value; path=/\r\n"+
			"\r\n", time.Now().UTC().Format(time.RFC1123))
	} else {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Date: %s\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"Foo: Bar\r\n"+
			"Transfer-Encoding: chunked\r\n"+
			"Foo: Baz\r\n"+
			"Set-Cookie: testcookie=value; path=/\r\n"+
			"\r\n", time.Now().UTC().Format(time.RFC1123))
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error writing response headers: %v\n", err)
		return
	}

	chunks := []string{
		"4\r\nTest\r\n",
		"A\r\nHelloWorld\r\n",
		"1\r\n\n\r\n",
		"0\r\n\r\n",
	}

	for _, chunk := range chunks {
		_, err := conn.Write([]byte(chunk))
		if err != nil {
			fmt.Printf("Error writing chunk: %v\n", err)
			return
		}
	}
}

func startServer(port string, duplicateTE bool) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error starting server on port %s: %v\n", port, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on :%s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Printf("%v: Connection from %s\n", port, conn.RemoteAddr())
		go handleConnection(conn, duplicateTE)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func startHealthCheckServer(port string) {
	http.HandleFunc("/", healthCheckHandler)
	fmt.Printf("Health check server is listening on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting health check server: %v\n", err)
	}
}

func main() {
	port1 := os.Getenv("SINGLE_TE_PORT")
	port2 := os.Getenv("DUPLICATE_TE_PORT")
	healthCheckPort := os.Getenv("HEALTH_PORT")

	if port1 == "" || port2 == "" || healthCheckPort == "" {
		fmt.Println("Environment variables SINGLE_TE_PORT, DUPLICATE_TE_PORT, and HEALTH_PORT must be set")
		return
	}

	// Start the first server (normal Transfer-Encoding).
	go startServer(port1, false)

	// Start the second server (duplicate Transfer-Encoding).
	go startServer(port2, true)

	// Start the health check server.
	go startHealthCheckServer(healthCheckPort)

	select {}
}
