// Package main implements a simple HTTP server that provides
// fine-grained control over the response headers, bypassing Go's
// built-in HTTP sanitization mechanisms. This is needed for testing
// scenarios where we need to return multiple "Transfer-Encoding"
// headers, which the standard Go HTTP stack would otherwise prevent
// by merging or overwriting duplicate TE headers.
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
	"time"
)

// connResponseWriter is a custom writer that handles HTTP response
// writing directly over a net.Conn. It provides methods to write
// headers and body to the connection, and it tracks any errors that
// occur during writing. This allows fine-grained control over the
// response, enabling scenarios that bypass Go's standard HTTP
// sanitisation mechanisms.
type connResponseWriter struct {
	conn     net.Conn
	header   http.Header
	writeErr error // Tracks if an error occurred during writing
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

// Header returns the HTTP headers.
func (w *connResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

// logWriteError records and logs an error that occurred while writing
// the response. It sets the writeErr field, captures the stack trace,
// truncates the logged content if it exceeds 200 characters, and logs
// the connection address, bytes written, truncated content, the
// error, and the caller's stack frames.
func (w *connResponseWriter) logWriteError(err error, bytesWritten int, content string) {
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

// Fprintf formats according to a format specifier and writes the
// resulting string to the connection. It logs and records any errors
// encountered during writing. If a previous write error exists,
// Fprintf returns immediately without writing. It returns the number
// of bytes written and any error encountered.
func (w *connResponseWriter) Fprintf(format string, a ...interface{}) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	s := fmt.Sprintf(format, a...)
	n, err := io.WriteString(w.conn, s)
	if err != nil {
		w.logWriteError(err, n, s)
	}
	return n, err
}

// Print writes the concatenated string representations of its
// arguments to the connection. If an error occurs during the write,
// it logs the error and updates the internal writeErr state. If a
// previous write error is recorded (writeErr is non-nil), Print will
// skip writing and return the existing error immediately. It returns
// the number of bytes written and any error encountered.
func (w *connResponseWriter) Print(a ...interface{}) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	s := fmt.Sprint(a...)
	n, err := io.WriteString(w.conn, s)
	if err != nil {
		w.logWriteError(err, n, s)
	}
	return n, err
}

// WriteHeader writes the HTTP status code and headers to the
// connection. It first writes the status line in the format "HTTP/1.1
// <statusCode> <statusText>\r\n". Then, it writes each header as
// "<key>: <value>\r\n". If any write operation encounters an error,
// it stops further writes and returns immediately. If a previous
// write error exists, WriteHeader does nothing.
func (w *connResponseWriter) WriteHeader(statusCode int) {
	if w.writeErr != nil {
		return
	}

	if _, err := w.Fprintf("HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode)); err != nil {
		return
	}

	for k, v := range w.Header() {
		for _, value := range v {
			if _, err := w.Fprintf("%s: %s\r\n", k, value); err != nil {
				return
			}
		}
	}

	if _, err := w.Print("\r\n"); err != nil {
		return
	}
}

// Write writes the byte slice 'b' directly to the underlying
// net.Conn. If an error occurs during writing, it logs the error and
// updates the internal writeErr state. If a previous write error
// exists, Write returns immediately without writing. It returns the
// number of bytes written and any error encountered during the write
// operation.
func (w *connResponseWriter) Write(b []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	n, err := w.conn.Write(b)
	if err != nil {
		w.logWriteError(err, n, string(b))
	}

	return n, err
}

// writeChunk writes a single chunk to the connection. If 'data' is
// non-empty, it writes the chunk size in hexadecimal followed by the
// data and a trailing CRLF. If 'data' is empty, it writes the
// terminating chunk ("0\r\n\r\n") to signal the end of the chunked
// transfer. It returns the number of bytes written and any error
// encountered during the write operation.
func (w *connResponseWriter) writeChunk(data string) (int, error) {
	if data == "" {
		return w.Print("0\r\n\r\n")
	}

	return w.Print(fmt.Sprintf("%X\r\n%s\r\n", len(data), data))
}

// setConnTimeout sets a deadline on the connection by adding the
// specified timeout to the current time. It returns a function to
// clear the timeout by resetting the connection's deadline to zero
// (i.e., no deadline). If setting the deadline fails, it returns an
// error. The returned function logs any error that occurs when
// clearing the deadline.
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

// handleConnection processes an incoming connection by setting a
// timeout, reading the request line, and routing it to the
// appropriate handler. It sets a connection timeout using
// setConnTimeout and logs any errors. Once the connection is
// processed or an error occurs, it ensures the connection is closed.
// The request is parsed for method, path, and protocol, and the
// request is routed based on the path.
func handleConnection(conn net.Conn) {
	log.Printf("%v: Connection from %v", conn.LocalAddr(), conn.RemoteAddr())

	clearTimeout, err := setConnTimeout(conn, 30*time.Second)
	if err != nil {
		log.Printf("set connection timeout failed: %v", err)
		conn.Close()
		return
	}

	writer := connResponseWriter{
		conn: conn,
	}

	defer func() {
		peer := conn.RemoteAddr()
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
// endpoint. For GET requests, it responds with an HTTP 200 OK status
// and writes "OK" to the response body. For HEAD requests, it
// responds with an HTTP 200 OK status without a body. If the request
// method is neither GET nor HEAD, it responds with an HTTP 405 Method
// Not Allowed status.
func healthzHandler(w *connResponseWriter, method string) {
	switch method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		w.Print("OK\n")
	case "HEAD":
		// Only write headers for HEAD requests.
		w.WriteHeader(http.StatusOK)
	default:
		// Method not allowed.
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Print(w, "405 Method Not Allowed\n")
	}
}

// handleSingleTE sends a response with a single Transfer-Encoding
// header. It sets the Date, Content-Type, Connection, and
// Transfer-Encoding headers, then writes an HTTP 200 OK status. The
// response body is sent in chunked encoding, with a single chunk
// containing "single-te\n" followed by the final chunk signaling the
// end of the response.
func handleSingleTE(w *connResponseWriter) {
	w.Header().Set("Date", time.Now().UTC().Format(time.RFC1123))
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Connection", "close")
	w.Header().Set("Transfer-Encoding", "chunked")

	w.WriteHeader(http.StatusOK)

	w.writeChunk("single-te\n")
	w.writeChunk("")
}

// handleDuplicateTE sends a response with duplicate Transfer-Encoding
// headers. It sets the Date, Content-Type, and Connection headers,
// and adds two "Transfer-Encoding: chunked" headers. The response
// status is set to HTTP 200 OK. The response body is sent in chunked
// encoding, with a single chunk containing "duplicate-te\n" followed
// by the final chunk signaling the end of the response.
func handleDuplicateTE(w *connResponseWriter) {
	w.Header().Set("Date", time.Now().UTC().Format(time.RFC1123))
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Connection", "close")
	w.Header().Add("Transfer-Encoding", "chunked")
	w.Header().Add("Transfer-Encoding", "chunked")

	w.WriteHeader(http.StatusOK)

	w.writeChunk("duplicate-te\n")
	w.writeChunk("")
}

// notFoundHandler responds with a 404 Not Found status and message.
// It sets the HTTP status to 404 and writes the body containing "404
// Not Found" in plain text.
func notFoundHandler(w *connResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Print("404 Not Found\n")
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
