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
	"encoding/json"
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
	"sync"
	"time"
)

// responseWriter is a custom writer that handles HTTP response
// writing directly over a net.Conn. It provides methods to write
// headers and body to the connection, and it tracks any errors that
// occur during writing. This allows fine-grained control over the
// response, enabling scenarios that bypass Go's standard HTTP
// sanitisation mechanisms.
type responseWriter struct {
	conn         net.Conn
	connWriteErr error
	httpHeader   http.Header
	logger       *responseLogger
}

type ConnectionLog struct {
	ConnError   string    `json:"conn_error,omitempty"`
	Entries     []string  `json:"entries"`
	LocalAddr   string    `json:"local_addr"`
	PeerAddr    string    `json:"peer_addr"`
	RequestLine string    `json:"request_line"`
	Timestamp   time.Time `json:"timestamp"`
}

type responseLogger struct {
	index int
}

const (
	logIndexSkip = -1
)

var (
	responseLogs   []ConnectionLog
	responseLogsMu sync.Mutex
)

func newResponseLogger(conn net.Conn) *responseLogger {
	responseLogsMu.Lock()
	defer responseLogsMu.Unlock()

	var peerAddr, localAddr string

	if conn.RemoteAddr() != nil {
		peerAddr = conn.RemoteAddr().String()
	} else {
		peerAddr = "unknown"
	}

	if conn.LocalAddr() != nil {
		localAddr = conn.LocalAddr().String()
	} else {
		localAddr = "unknown"
	}

	index := len(responseLogs)

	responseLogs = append(responseLogs, ConnectionLog{
		PeerAddr:  peerAddr,
		LocalAddr: localAddr,
		Timestamp: time.Now(),
	})

	return &responseLogger{index: index}
}

func (rl *responseLogger) setRequestLine(requestLine string) {
	responseLogsMu.Lock()
	defer responseLogsMu.Unlock()

	responseLogs[rl.index].RequestLine = requestLine

	log.Printf("%v >>> %v Received request: %s",
		responseLogs[rl.index].LocalAddr,
		responseLogs[rl.index].PeerAddr,
		responseLogs[rl.index].RequestLine)
}

func (rl *responseLogger) logResponse(data string) {
	if rl.index == logIndexSkip {
		return
	}

	for _, msg := range rl.formatResponseData(data) {
		rl.recordMessage(msg)
	}
}

func (rl *responseLogger) logError(err error, bytesWritten int, content string) {
	if rl.index == logIndexSkip {
		return
	}

	if err != nil {
		rl.recordError(err)
	}

	for _, msg := range rl.formatWriteError(err, bytesWritten, content) {
		rl.recordMessage(msg)
	}
}

func (rl *responseLogger) formatResponseData(data string) []string {
	var result []string
	maxContentLength := 200

	if len(data) > maxContentLength {
		data = data[:maxContentLength] + "... (truncated)"
	}

	if strings.HasPrefix(data, "HTTP/") || strings.Contains(data, ": ") {
		result = append(result, data)
		return result
	}

	if strings.HasSuffix(data, "\r\n") {
		parts := strings.SplitN(data, "\r\n", 2)
		if len(parts) == 2 {
			chunkSize := parts[0]
			chunkData := strings.TrimSuffix(parts[1], "\r\n")

			if chunkSize != "" {
				result = append(result, "Chunk-Size: 0x"+chunkSize)
			}

			if chunkData != "" {
				result = append(result, "Chunk-Data: "+chunkData)
			}

			if chunkSize == "0" {
				result = append(result, "End-Of-Chunked-Transfer")
			}

			return result
		}
	}

	if data == "" {
		result = append(result, "Empty-Response")
		return result
	}

	result = append(result, "Body: "+strings.TrimSpace(data))
	return result
}

func (rl *responseLogger) formatWriteError(err error, bytesWritten int, content string) []string {
	maxContentLength := 200
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "... (truncated)"
	}

	var result = []string{
		fmt.Sprintf("error writing %q after %d bytes: %v", content, bytesWritten, err),
	}

	return append(result, formatStackTrace(getStackFrames(6))...)
}

func (rl *responseLogger) recordError(err error) {
	responseLogsMu.Lock()
	defer responseLogsMu.Unlock()

	if responseLogs[rl.index].ConnError == "" && err != nil {
		responseLogs[rl.index].ConnError = err.Error()
	}
}

func (rl *responseLogger) recordMessage(format string, args ...interface{}) {
	if rl.index == logIndexSkip {
		return
	}

	responseLogsMu.Lock()
	defer responseLogsMu.Unlock()

	msg := fmt.Sprintf(format, args...)
	responseLogs[rl.index].Entries = append(responseLogs[rl.index].Entries, msg)

	log.Printf("%v <<< %v %s",
		responseLogs[rl.index].LocalAddr,
		responseLogs[rl.index].PeerAddr,
		msg)

}

func (rl *responseLogger) skipLogging() {
	rl.index = logIndexSkip
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

func (w *responseWriter) logResponseData(data string) {
	w.logger.logResponse(data)
}

func (w *responseWriter) logWriteError(err error, bytesWritten int, content string) {
	w.logger.logError(err, bytesWritten, content)
}

// formatStackTrace converts a slice of runtime.Frame into a slice of
// formatted strings. Each string represents a stack frame with the
// function name, file name, and line number.
func formatStackTrace(frames []runtime.Frame) []string {
	var stackElements []string

	for _, frame := range frames {
		file := filepath.Base(frame.File)
		stackElements = append(stackElements, fmt.Sprintf("%s (%s:%d)", frame.Function, file, frame.Line))
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

func (w *responseWriter) writeToConnection(data string) (int, error) {
	if w.connWriteErr != nil {
		return 0, w.connWriteErr
	}

	n, err := io.WriteString(w.conn, data)
	if err != nil {
		w.logWriteError(err, n, data)
	} else {
		w.logResponseData(data)
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
	return w.writeToConnection(string(b))
}

// writeHeader writes the HTTP status code and headers to the
// connection. It first writes the status line in the format "HTTP/1.1
// <statusCode> <statusText>\r\n". Then, it writes each header as
// "<key>: <value>\r\n". If any write operation encounters an error,
// it stops further writes and returns immediately. If a previous
// write error exists, writeHeader does nothing.
func (w *responseWriter) writeHeader(statusCode int) {
	if w.connWriteErr != nil {
		return
	}

	w.writeToConnection(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode)))

	var keys []string
	for k := range w.header() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range w.header()[k] {
			w.writeToConnection(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}

	w.writeToConnection("\r\n")
}

// writeChunk writes a single chunk to the connection. If 'data' is
// non-empty, it writes the chunk size in hexadecimal followed by the
// data and a trailing CRLF. If 'data' is empty, it writes the
// terminating chunk ("0\r\n\r\n") to signal the end of the chunked
// transfer. It returns the number of bytes written and any error
// encountered during the write operation.
func (w *responseWriter) writeChunk(data string) (int, error) {
	if data == "" {
		return w.writeToConnection("0\r\n\r\n")
	}

	return w.writeToConnection(fmt.Sprintf("%X\r\n%s\r\n", len(data), data))
}

// setConnTimeout sets a deadline on the connection by adding the
// specified timeout to the current time. It returns a function to
// clear the timeout by resetting the connection's deadline to zero
// (i.e., no deadline). If setting the deadline fails, it returns an
// error. The returned function logs any error that occurs when
// clearing the deadline.
func (w *responseWriter) setConnTimeout(timeout time.Duration) (clearTimeout func() error, err error) {
	if err := w.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	return func() error {
		return w.conn.SetDeadline(time.Time{})
	}, nil
}

// Error replies to the request with the specified error message and
// HTTP code. It does not otherwise end the request; the caller should
// ensure no further writes are done to w. The error message should be
// plain text.
func (w *responseWriter) httpError(error string, code int) {
	w.header().Set("Content-Type", "text/plain; charset=utf-8")
	w.header().Set("X-Content-Type-Options", "nosniff")
	w.writeHeader(code)
	w.write([]byte(error))
}

// handleConnection processes an incoming connection by setting a
// timeout, reading the request line, and routing it to the
// appropriate handler. It sets a connection timeout using
// setConnTimeout and logs any errors. Once the connection is
// processed or an error occurs, it ensures the connection is closed.
// The request is parsed for method, path, and protocol, and the
// request is routed based on the path.
func handleConnection(conn net.Conn) {
	logger := newResponseLogger(conn)

	writer := responseWriter{
		conn:   conn,
		logger: logger,
	}

	clearTimeout, err := writer.setConnTimeout(30 * time.Second)
	if err != nil {
		writer.httpError("Internal Server Error", http.StatusInternalServerError)
		conn.Close()
		return
	}

	defer func() {
		timeoutErr := clearTimeout()
		closeErr := conn.Close()
		if timeoutErr != nil || closeErr != nil {
			logger.recordMessage("Error during connection cleanup: %v %v", timeoutErr, closeErr)
		}
	}()

	reader := bufio.NewReader(conn)
	requestLine, _, err := reader.ReadLine()
	if err != nil {
		writer.httpError(fmt.Sprintf("Failed to read request line: %v", err), http.StatusBadRequest)
		return
	}

	logger.setRequestLine(string(requestLine))

	var method, path, protocol string
	if _, err = fmt.Sscanf(string(requestLine), "%s %s %s", &method, &path, &protocol); err != nil {
		writer.httpError("Error parsing request line", http.StatusBadRequest)
		return
	}

	switch path {
	case "/healthz":
		healthzHandler(&writer, method)
	case "/single-te":
		handleSingleTransferEncodingRequest(&writer, method)
	case "/duplicate-te":
		handleDuplicateTransferEncodingRequest(&writer, method)
	case "/access-logs":
		accessLogsHandler(&writer, method)
	default:
		writer.httpError("Not found\n", http.StatusNotFound)
	}
}

// handleHTTPRequest is a comprehensive handler for various HTTP
// endpoints. It supports different status codes, chunked and
// non-chunked responses, and handles GET, HEAD, and other HTTP
// methods consistently.
//
// Behavior:
//   - GET: Sends the full response with headers and body.
//   - HEAD: Sends only the headers without a body, but includes Content-Length if applicable.
//   - Other methods: Responds with the specified status code, but does not send a body.
func handleHTTPRequest(w *responseWriter, method string, statusCode int, body string, useChunkedEncoding bool, additionalTEHeaders int) {
	// Set basic headers.
	w.header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	w.header().Set("Content-Type", "text/plain")

	// Determine if a body can or should be sent (skip for 1xx,
	// 204, 304).
	bodyAllowed := statusCode >= 200 && statusCode != http.StatusNoContent && statusCode != http.StatusNotModified

	// Handle HEAD requests: set Content-Length, but skip sending
	// the body.
	if method == http.MethodHead {
		if bodyAllowed {
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		}
		w.writeHeader(statusCode)
		// No body is sent for HEAD requests.
		return
	}

	// For GET or other methods, handle Transfer-Encoding and
	// Content-Length.
	if bodyAllowed {
		if useChunkedEncoding {
			w.header().Set("Transfer-Encoding", "chunked")
			for i := 0; i < additionalTEHeaders; i++ {
				w.header().Add("Transfer-Encoding", "chunked")
			}
		} else {
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		}
	}

	// Write the headers to the client.
	w.writeHeader(statusCode)

	// Send the body only for GET and other non-HEAD methods.
	if bodyAllowed && method == http.MethodGet {
		if useChunkedEncoding {
			w.writeChunk(body)
			w.writeChunk("") // End chunked transfer
		} else {
			w.write([]byte(body)) // Send entire body
		}
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

func accessLogsHandler(w *responseWriter, method string) {
	// Skip logging for this request to avoid recursive logging.
	w.logger.skipLogging()

	responseLogsMu.Lock()
	logs := make([]ConnectionLog, len(responseLogs))
	copy(logs, responseLogs)
	responseLogsMu.Unlock()

	logsJSON, err := json.Marshal(logs)
	if err != nil {
		w.httpError("Failed to marshal access logs", http.StatusInternalServerError)
		return
	}

	handleHTTPRequest(w, method, http.StatusOK, string(logsJSON), false, 0)
}

// createListener creates a TCP listener on the specified port. If
// useTLS is true, it loads a TLS certificate and configures the
// listener for secure connections. It returns a net.Listener or an
// error if the listener creation fails.
func createListener(port string, useTLS bool) (net.Listener, error) {
	if useTLS {
		cert, err := tls.LoadX509KeyPair("/etc/serving-cert/tls.crt", "/etc/serving-cert/tls.key")
		if err != nil {
			log.Printf("Error loading TLS certificate: %v\n", err)
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
