package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
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
func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("%v: Connection from %s\n", conn.LocalAddr(), conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read the request line (e.g., "GET /duplicate-te HTTP/1.1").
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("%v: Error reading request line: %v\n", conn.LocalAddr(), err)
		return
	}

	log.Printf("%v: Received request: %s", conn.LocalAddr(), requestLine)

	// Split the request line to get the method, path, and protocol.
	var method, path, protocol string
	_, err = fmt.Sscanf(requestLine, "%s %s %s", &method, &path, &protocol)
	if err != nil {
		log.Printf("%v: Error parsing request line: %v\n", conn.LocalAddr(), err)
		return
	}

	switch path {
	case "/healthz":
		healthResponse := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Connection: close\r\n\r\n" +
			"OK\r\n"
		if _, err := writer.WriteString(healthResponse); err != nil {
			log.Printf("%v: Error writing /healthz response: %v\n", conn.LocalAddr(), err)
		}
		writer.Flush()
		return

	case "/single-te":
		handleSingleTE(conn, writer)

	case "/duplicate-te":
		handleDuplicateTE(conn, writer)

	default:
		notFoundResponse := "HTTP/1.1 404 Not Found\r\n" +
			"Content-Type: text/plain\r\n" +
			"Connection: close\r\n\r\n" +
			"404 page not found\r\n"
		if _, err := writer.WriteString(notFoundResponse); err != nil {
			log.Printf("%v: Error writing 404 response: %v\n", conn.LocalAddr(), err)
		}
		writer.Flush()
		return
	}

	if err := writer.Flush(); err != nil {
		log.Printf("%v: Error flushing data to connection: %v\n", conn.LocalAddr(), err)
	}
}

// handleSingleTE handles the response for the /single-te path.
func handleSingleTE(conn net.Conn, writer *bufio.Writer) {
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Date: %s\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"Connection: close\r\n"+
		"Transfer-Encoding: chunked\r\n\r\n", time.Now().UTC().Format(time.RFC1123))

	if _, err := writer.WriteString(response); err != nil {
		log.Printf("%v: Error writing single-te response headers: %v\n", conn.LocalAddr(), err)
		return
	}

	// Form the chunked response
	chunks := []string{
		// First chunk:
		// - 'A' is the hexadecimal representation of 10 (the length of "single-te\n").
		// - \r\n separates the chunk size from the chunk data.
		// - "single-te\n" is the chunk data (the \n is included in the chunk data).
		// - \r\n terminates the chunk.
		"A\r\nsingle-te\n\r\n",

		// Final chunk:
		// - '0' indicates a chunk of zero length, signaling the end of the chunked message.
		// - The first \r\n separates the chunk size (0) from what would be the chunk data.
		// - The second \r\n terminates the zero-length chunk and the entire chunked message.
		"0\r\n\r\n",
	}

	for i := range chunks {
		if _, err := writer.WriteString(chunks[i]); err != nil {
			log.Printf("%v: Error writing chunk: %v\n", conn.LocalAddr(), err)
			break
		}
	}
}

// handleDuplicateTE handles the response for the /duplicate-te path.
func handleDuplicateTE(conn net.Conn, writer *bufio.Writer) {
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Date: %s\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"Connection: close\r\n"+
		"Transfer-Encoding: chunked\r\n"+
		"Transfer-Encoding: chunked\r\n\r\n", time.Now().UTC().Format(time.RFC1123))

	if _, err := writer.WriteString(response); err != nil {
		log.Printf("%v: Error writing duplicate-te response headers: %v\n", conn.LocalAddr(), err)
		return
	}

	chunks := []string{
		// First chunk:
		// - 'D' is the hexadecimal representation of 13 (the length of "duplicate-te\n").
		// - \r\n separates the chunk size from the chunk data.
		// - "duplicate-te\n" is the chunk data (note the \n is part of the data).
		// - \r\n terminates the chunk.
		"D\r\nduplicate-te\n\r\n",

		// Final chunk:
		// - '0' indicates a chunk of zero length, signaling the end of the chunked message.
		// - The first \r\n separates the chunk size (0) from what would be the chunk data.
		// - The second \r\n terminates the zero-length chunk and the entire chunked message.
		"0\r\n\r\n",
	}

	for i := range chunks {
		if _, err := writer.WriteString(chunks[i]); err != nil {
			log.Printf("%v: Error writing chunk: %v\n", conn.LocalAddr(), err)
			break
		}
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

func startServer(port string, useTLS bool) {
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
		go handleConnection(conn)
	}
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	httpsPort := os.Getenv("HTTPS_PORT")

	if httpPort == "" || httpsPort == "" {
		log.Fatalf("Environment variables HTTP_PORT and HTTPS_PORT must be set\n")
		os.Exit(1)
	}

	// Start non-TLS server handling all paths, including /healthz
	go startServer(httpPort, false)

	// Start TLS server handling all paths, including /healthz
	go startServer(httpsPort, true)

	// Block forever (needed to keep the servers running).
	select {}
}
