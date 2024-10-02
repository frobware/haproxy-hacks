// Package main implements a simple HTTP server that provides
// fine-grained control over the response headers and body, bypassing
// Go's built-in HTTP sanitisation mechanisms. This is particularly
// useful for testing scenarios where we need to return multiple
// "Transfer-Encoding" headers, which the standard Go HTTP stack would
// otherwise reject.
//
// The key design involves a custom implementation of the
// http.ResponseWriter interface using a raw net.Conn, which allows
// manual handling of headers, status codes, and body chunks. The
// server supports flushing chunks directly to the client, which is
// essential for simulating chunked transfer encoding.
//
// The connResponseWriter struct is central to the design, managing
// the connection, headers, status, and any write errors. It buffers
// writes to the connection using bufio.Writer, allowing headers and
// body chunks to be written out efficiently and then flushed to the
// client.
//
// Key design points include:
//
// 1. Fine-Grained Control Over HTTP Responses:
//
//   - The server allows manual control over headers, status codes,
//     and the response body, enabling scenarios where multiple
//     headers (e.g., Transfer-Encoding) need to be set.
//
//   - This custom implementation circumvents Go’s default
//     sanitisation, which merges headers, making it particularly
//     useful for testing edge cases.
//
// 2. Simplified Error Handling for Handlers:
//
//   - One of the key features is the tracking of write and flush
//     errors in the connResponseWriter. If a write or flush fails,
//     subsequent write calls become cheap no-ops, making future
//     operations safe and error-free.
//
//   - This design choice significantly simplifies the handler logic,
//     as handlers can focus on writing headers and body content
//     without repetitive error handling code. The handlers remain
//     explicit and free from concerns about past write errors.
//
// 3. Chunked Transfer Encoding Support:
//
//   - Functions like writeChunk and writeChunks allow the server to
//     simulate chunked transfer encoding, flushing chunks to the
//     client as they are written.
//
//   - This design enables handling of different paths (/single-te,
//     /duplicate-te) where different headers and transfer encodings
//     are returned.
//
// 4. Connection and Error Management:
//
//   - The connResponseWriter tracks any errors during writes and
//     ensures that subsequent writes are skipped once an error
//     occurs, reducing the likelihood of multiple errors.

//   - The flush mechanism ensures that buffered data is sent to the
//     client immediately when needed.
//
// 5. TLS and Non-TLS Support:
//
//   - The server supports both plain TCP connections and TLS-secured
//     connections, based on the configuration provided at startup.
//
// 6. Graceful Shutdown:
//
//   - The server listens for SIGINT and SIGTERM signals and shuts
//     down gracefully when a termination signal is received, ensuring
//     proper cleanup of connections and resources.
//
// Endpoints:
//
//   - /healthz: Returns a simple health check response.
//
//   - /single-te: Returns a response with a single
//     "Transfer-Encoding" header and uses chunked transfer encoding.
//
//   - /duplicate-te: Returns a response with multiple
//     "Transfer-Encoding" headers, simulating an invalid but testable
//     scenario.
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// connResponseWriter implements http.ResponseWriter using a raw
// net.Conn and adds the ability to flush chunks to the client
// immediately using a buffered writer.
type connResponseWriter struct {
	conn     net.Conn
	writer   *bufio.Writer // Buffers the output to conn
	header   http.Header
	status   int
	writeErr error // Tracks if an error occurred during writing
}

// Header returns the HTTP headers.
func (w *connResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

// WriteHeader writes the HTTP status code manually to the raw
// connection.
func (w *connResponseWriter) WriteHeader(statusCode int) {
	if w.writeErr != nil {
		// Skip future writes if there's already been a write
		// error.
		return
	}
	w.status = statusCode
	_, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))
	if err != nil {
		w.writeErr = err
		return
	}
	for k, v := range w.Header() {
		for _, value := range v {
			_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, value)
			if err != nil {
				w.writeErr = err
				return
			}
		}
	}
	_, err = fmt.Fprint(w.writer, "\r\n")
	if err != nil {
		w.writeErr = err
	}
}

// Write writes the response body to the connection.
func (w *connResponseWriter) Write(b []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr // Skip if there's already been a write error
	}
	n, err := w.writer.Write(b)
	if err != nil {
		w.writeErr = err
	}
	return n, err
}

// Flush flushes any buffered data to the client.
func (w *connResponseWriter) Flush() {
	if w.writeErr != nil {
		// Skip if there's already been a write error.
		return
	}
	err := w.writer.Flush()
	if err != nil {
		w.writeErr = err
	}
}

// writeChunk writes a single chunk to the provided writer and flushes
// it.
func writeChunk(w *connResponseWriter, chunk string) {
	w.Write([]byte(chunk))
	w.Flush()
}

// writeChunks writes multiple chunks to the provided writer.
func writeChunks(w *connResponseWriter, chunks []string) {
	for i := range chunks {
		writeChunk(w, chunks[i])
	}
}

// setConnTimeout sets a timeout on the given connection and returns a
// function to clear the timeout.
func setConnTimeout(conn net.Conn, timeout time.Duration) (clearTimeout func(), err error) {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("error setting connection deadline: %w", err)
	}

	return func() {
		if err := conn.SetDeadline(time.Time{}); err != nil {
			log.Printf("Error clearing connection deadline: %v", err)
		}
	}, nil
}

// handleConnection processes incoming connections and routes to
// different handlers.
func handleConnection(conn net.Conn) {
	log.Printf("%v: Connection from %v", conn.LocalAddr(), conn.RemoteAddr())

	clearTimeout, err := setConnTimeout(conn, 30*time.Second)
	if err != nil {
		log.Printf("%v", err)
		conn.Close()
		return
	}

	writer := connResponseWriter{
		conn:   conn,
		writer: bufio.NewWriter(conn),
	}

	defer func() {
		peer := conn.RemoteAddr()
		writer.Flush()
		clearTimeout()
		conn.Close()
		log.Printf("%v: Closed connection for %v", conn.LocalAddr(), peer)
	}()

	reader := bufio.NewReader(conn)
	requestLine, _, err := reader.ReadLine()
	if err != nil {
		log.Printf("%v: Error reading request line: %v\n", conn.LocalAddr(), err)
		return
	}
	log.Printf("%v: Received request: %s", conn.LocalAddr(), requestLine)

	// Parse the request line (method, path, protocol).
	var method, path, protocol string
	if _, err = fmt.Sscanf(string(requestLine), "%s %s %s", &method, &path, &protocol); err != nil {
		log.Printf("%v: Error parsing request line: %v\n", conn.LocalAddr(), err)
		return
	}

	switch path {
	case "/healthz":
		healthzHandler(&writer, method)
	case "/single-te":
		handleSingleTE(&writer)
	case "/duplicate-te":
		handleDuplicateTE(&writer)
	default:
		notFoundHandler(&writer)
	}
}

// healthzHandler handles both GET and HEAD requests for the /healthz
// endpoint.
func healthzHandler(w *connResponseWriter, method string) {
	switch method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	case "HEAD":
		// Only write headers for HEAD requests
		w.WriteHeader(http.StatusOK)
	default:
		// Method not allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "405 Method Not Allowed")
	}
}

// handleSingleTE sends a response with a single Transfer-Encoding header using Header().
func handleSingleTE(w *connResponseWriter) {
	w.Header().Set("Date", time.Now().UTC().Format(time.RFC1123))
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Connection", "close")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Write the headers and status code
	w.WriteHeader(http.StatusOK)

	// Write chunks using helper functions
	writeChunks(w, []string{
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
	})
}

// handleDuplicateTE sends a response with duplicate Transfer-Encoding
// headers using Header().
func handleDuplicateTE(w *connResponseWriter) {
	w.Header().Set("Date", time.Now().UTC().Format(time.RFC1123))
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Connection", "close")
	w.Header().Add("Transfer-Encoding", "chunked")
	w.Header().Add("Transfer-Encoding", "chunked")

	w.WriteHeader(http.StatusOK)

	writeChunks(w, []string{
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
	})
}

// notFoundHandler responds with 404 Not Found.
func notFoundHandler(w *connResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 Not Found")
	w.Flush()
}

// createListener creates a TCP listener with optional TLS.
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

// startServer starts the server on the given port with or without
// TLS.
func startServer(port string, useTLS bool) {
	listener, err := createListener(port, useTLS)
	if err != nil {
		log.Fatalf("Error starting server on port %s: %v\n", port, err)
	}
	defer listener.Close()

	log.Printf("Server is listening on port %s (TLS=%v)\n", port, useTLS)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
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
	}

	go startServer(httpPort, false) // Start non-TLS server.
	go startServer(httpsPort, true) // Start TLS server.

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Termination signal received, exiting...")
	os.Exit(0)
}
