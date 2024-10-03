// Package main implements an HTTP server with fine-grained control
// over response headers, bypassing Go's HTTP sanitisation for testing
// scenarios with multiple "Transfer-Encoding" headers.
//
// It introduces connResponseWriter, a custom response writer that:
// - Centralises logging of headers, body content, and write errors
// - Maintains internal error state to prevent cascading failures
//
// This enables focused response logic in handlers without nil checks,
// while providing detailed debugging information.
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// responseWriter is a custom writer that handles HTTP response
// writing directly over a net.Conn. It provides methods to write
// headers and body to the connection, and it tracks any errors that
// occur during writing. This allows fine-grained control over the
// response, enabling scenarios that bypass Go's standard HTTP
// sanitisation mechanisms.
type responseWriter struct {
	bodySize   int64
	conn       net.Conn
	httpHeader http.Header
	isChunked  bool
	writeErr   error // Tracks if an error occurred during writing
}

// getStackFrames returns a slice of runtime.Frame, skipping the
// specified number of frames.
func getStackFrames(skip int) []runtime.Frame {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	var stackFrames []runtime.Frame

	for {
		frame, more := frames.Next()
		stackFrames = append(stackFrames, frame)
		if !more {
			break
		}
	}

	return stackFrames
}

// logResponseData logs the response data with the following behavior:
//
// - Truncates content longer than 200 characters
//
// - Logs headers as-is
//
// - For chunked data:
//   - Logs chunk size (if present)
//   - Logs chunk data (if present)
//   - Logs end of chunked transfer
//
// - Logs non-chunked body content
//
// - Logs empty responses
func (w *responseWriter) logResponseData(data string) {
	maxContentLength := 200

	// Truncate the content if it's too long.
	if len(data) > maxContentLength {
		data = data[:maxContentLength] + "... (truncated)"
	}

	prefix := fmt.Sprintf("%v: Writing to %v: ", w.conn.LocalAddr(), w.conn.RemoteAddr())

	// Handle headers.
	if strings.HasPrefix(data, "HTTP/") || strings.Contains(data, ": ") {
		log.Printf(prefix + data)
		return
	}

	// Handle chunked data.
	if strings.HasSuffix(data, "\r\n") {
		parts := strings.SplitN(data, "\r\n", 2)
		if len(parts) == 2 {
			chunkSize := parts[0]
			chunkData := strings.TrimSuffix(parts[1], "\r\n")

			if chunkSize != "" {
				log.Printf(prefix + "Chunk-Size: 0x" + chunkSize)
			}

			if chunkData != "" {
				log.Printf(prefix + "Chunk-Data: " + chunkData)
			}

			if chunkSize == "0" {
				log.Printf(prefix + "End-Of-Chunked-Transfer")
			}

			return
		}
	}

	if data == "" {
		log.Printf(prefix + "Empty-Response")
		return
	}

	// Handle any other data (non-chunked body content).
	log.Printf(prefix + "Body: " + strings.TrimSpace(data))
}

// formatStackTrace converts a slice of runtime.Frame into a slice of
// formatted strings. Each string represents a stack frame with the
// function name, file name, and line number.
func formatStackTrace(frames []runtime.Frame) []string {
	var stackElements []string
	for _, frame := range frames {
		file := filepath.Base(frame.File)
		functionName := frame.Function
		element := fmt.Sprintf("%s (%s:%d)", functionName, file, frame.Line)
		stackElements = append(stackElements, element)
	}

	if len(stackElements) == 0 {
		return nil
	}

	return stackElements
}

// header returns the HTTP headers.
func (w *responseWriter) header() http.Header {
	if w.httpHeader == nil {
		w.httpHeader = make(http.Header)
	}
	return w.httpHeader
}

// logWriteError records and logs an error that occurred while writing
// the response. It sets the writeErr field, captures the stack trace,
// truncates the logged content if it exceeds 200 characters, and logs
// the connection address, bytes written, truncated content, the
// error, and the caller's stack frames.
func (w *responseWriter) logWriteError(err error, bytesWritten int, content string) {
	w.writeErr = err
	frames := getStackFrames(3)

	maxContentLength := 200
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "... (truncated)"
	}

	log.Printf("%v: error writing response after %d bytes. Content: %q, Error: %v. Caller:",
		w.conn.LocalAddr(), bytesWritten, content, err)

	for _, frame := range formatStackTrace(frames) {
		log.Printf("\t%s", frame)
	}
}

// fprintf formats according to a format specifier and writes the
// resulting string to the connection. It logs and records any errors
// encountered during writing. If a previous write error exists,
// fprintf returns immediately without writing. It returns the number
// of bytes written and any error encountered.
func (w *responseWriter) fprintf(format string, a ...interface{}) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	s := fmt.Sprintf(format, a...)
	w.logResponseData(s)

	n, err := io.WriteString(w.conn, s)
	if err != nil {
		w.logWriteError(err, n, s)
	}
	return n, err
}

// print writes the concatenated string representations of its
// arguments to the connection. If an error occurs during the write,
// it logs the error and updates the internal writeErr state. If a
// previous write error is recorded (writeErr is non-nil), print will
// skip writing and return the existing error immediately. It returns
// the number of bytes written and any error encountered.
func (w *responseWriter) print(a ...interface{}) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	s := fmt.Sprint(a...)
	w.logResponseData(s)

	n, err := io.WriteString(w.conn, s)
	if err != nil {
		w.logWriteError(err, n, s)
	}
	return n, err
}

// write writes the byte slice 'b' directly to the underlying
// net.Conn. If an error occurs during writing, it logs the error and
// updates the internal writeErr state. If a previous write error
// exists, write returns immediately without writing. It returns the
// number of bytes written and any error encountered during the write
// operation.
func (w *responseWriter) write(b []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	w.logResponseData(string(b))

	// Only set Content-Length if it's not already set and we're
	// not using chunked encoding.
	if !w.isChunked && w.header().Get("Content-Length") == "" && w.bodySize > 0 {
		w.header().Set("Content-Length", strconv.FormatInt(w.bodySize, 10))
	}

	n, err := w.conn.Write(b)
	if err != nil {
		w.logWriteError(err, n, string(b))
	}

	return n, err
}

// writeHeader writes the HTTP status code and headers to the
// connection. It first writes the status line in the format "HTTP/1.1
// <statusCode> <statusText>\r\n". Then, it writes each header as
// "<key>: <value>\r\n". If any write operation encounters an error,
// it stops further writes and returns immediately. If a previous
// write error exists, writeHeader does nothing.
func (w *responseWriter) writeHeader(statusCode int) {
	if w.writeErr != nil {
		return
	}

	if !w.isChunked && w.header().Get("Content-Length") == "" {
		w.header().Set("Content-Length", strconv.FormatInt(w.bodySize, 10))
	}

	if _, err := w.fprintf("HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode)); err != nil {
		return
	}

	var keys []string
	for k := range w.header() {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		for _, v := range w.header()[k] {
			if _, err := w.fprintf("%s: %s\r\n", k, v); err != nil {
				return
			}
		}
	}

	if _, err := w.print("\r\n"); err != nil {
		return
	}
}

// writeChunk writes a single chunk to the connection. If 'data' is
// non-empty, it writes the chunk size in hexadecimal followed by the
// data and a trailing CRLF. If 'data' is empty, it writes the
// terminating chunk ("0\r\n\r\n") to signal the end of the chunked
// transfer. It returns the number of bytes written and any error
// encountered during the write operation.
func (w *responseWriter) writeChunk(data string) (int, error) {
	if data == "" {
		return w.print("0\r\n\r\n")
	}

	return w.print(fmt.Sprintf("%X\r\n%s\r\n", len(data), data))
}

// setConnTimeout sets a deadline on the connection by adding the
// specified timeout to the current time. It returns a function to
// clear the timeout by resetting the connection's deadline to zero
// (i.e., no deadline). If setting the deadline fails, it returns an
// error. The returned function logs any error that occurs when
// clearing the deadline.
func (w *responseWriter) setConnTimeout(timeout time.Duration) (clearTimeout func(), err error) {
	if err := w.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("error setting connection deadline: %w", err)
	}

	return func() {
		if err := w.conn.SetDeadline(time.Time{}); err != nil {
			log.Printf("Error clearing connection deadline: %v", err)
		}
	}, nil
}

// handleConnection processes an incoming connection by setting a
// timeout, reading the request line, and routing it to the
// appropriate handler. It sets a connection timeout using
// setConnTimeout and logs any errors. Once the connection is
// processed or an error occurs, it ensures the connection is closed.
// The request is parsed for method, path, and protocol, and the
// request is routed based on the path.
func handleConnection(conn net.Conn) {
	log.Printf("----------------------------------------")
	log.Printf("%v: Connection from %v", conn.LocalAddr(), conn.RemoteAddr())

	writer := responseWriter{
		conn: conn,
	}

	clearTimeout, err := writer.setConnTimeout(30 * time.Second)
	if err != nil {
		log.Printf("set connection timeout failed: %v", err)
		conn.Close()
		return
	}

	closeConn := func() {
		clearTimeout()
		conn.Close()
		log.Printf("%v: Closed connection for %v", conn.LocalAddr(), conn.RemoteAddr())
	}
	defer closeConn()

	reader := bufio.NewReader(conn)
	requestLine, _, err := reader.ReadLine()
	if err != nil {
		log.Printf("%v: Error reading request line: %v\n", conn.LocalAddr(), err)
		return
	}

	log.Printf("%v: Received request: %s", conn.LocalAddr(), requestLine)

	var method, path, protocol string
	if _, err = fmt.Sscanf(string(requestLine), "%s %s %s", &method, &path, &protocol); err != nil {
		log.Printf("%v: Error parsing request line: %v\n", conn.LocalAddr(), err)
		return
	}

	switch path {
	case "/healthz":
		healthzHandler(&writer, method)
	case "/single-te":
		handleSingleTransferEncodingRequest(&writer, method)
	case "/duplicate-te":
		handleDuplicateTransferEncodingRequest(&writer, method)
	default:
		notFoundHandler(&writer, method)
	}
}

// handleHTTPRequest is a comprehensive handler for various HTTP
// endpoints. It supports different status codes, chunked and
// non-chunked responses, and handles GET, HEAD, and other HTTP
// methods consistently.
//
// Behavior:
//   - GET: Sends the full response with headers and body.
//   - HEAD: Sends only the headers that would be sent for a GET request.
//   - Other methods: Responds based on the provided status code and body.
func handleHTTPRequest(w *responseWriter, method string, statusCode int, body string, useChunkedEncoding bool, additionalTEHeaders int) {
	w.header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	w.header().Set("Content-Type", "text/plain")

	switch method {
	case http.MethodGet, http.MethodHead:
		if method == http.MethodHead {
			// For HEAD requests, always set Content-Length and never use chunked encoding
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		} else if useChunkedEncoding {
			w.header().Set("Transfer-Encoding", "chunked")
			for i := 0; i < additionalTEHeaders; i++ {
				w.header().Add("Transfer-Encoding", "chunked")
			}
			w.isChunked = true
		} else {
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		}
		w.writeHeader(statusCode)
		if method == http.MethodGet {
			if w.isChunked {
				w.writeChunk(body)
				w.writeChunk("")
			} else {
				w.print(body)
			}
		}
	default:
		// For methods other than GET and HEAD.
		notAllowedBody := "405 Method Not Allowed\n"
		w.header().Set("Content-Length", strconv.Itoa(len(notAllowedBody)))
		w.writeHeader(http.StatusMethodNotAllowed)
		w.print(notAllowedBody)
	}
}

func healthzHandler(w *responseWriter, method string) {
	handleHTTPRequest(w, method, http.StatusOK, "OK\n", false, 0)
}

func handleSingleTransferEncodingRequest(w *responseWriter, method string) {
	handleHTTPRequest(w, method, http.StatusOK, "single-te\n", true, 0)
}

func handleDuplicateTransferEncodingRequest(w *responseWriter, method string) {
	handleHTTPRequest(w, method, http.StatusOK, "duplicate-te\n", true, 1)
}

func notFoundHandler(w *responseWriter, method string) {
	handleHTTPRequest(w, method, http.StatusNotFound, "404 Not Found\n", false, 0)
}

// createListener creates a TCP listener on the specified port. If
// useTLS is true, it loads a TLS certificate and configures the
// listener for secure connections. It returns a net.Listener or an
// error if the listener creation fails.
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

// startServer starts a TCP server on the specified port with optional
// TLS. It creates a listener using the createListener function, which
// may configure TLS if useTLS is true. The server listens for
// incoming connections and spawns a new goroutine for each connection
// using handleConnection. If an error occurs during the creation of
// the listener, the function logs a fatal error and exits.
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

	select {}
}
