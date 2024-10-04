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
	logger       *ResponseLogger
}

type ConnectionLog struct {
	ConnError   string    `json:"conn_error,omitempty"`
	Entries     []string  `json:"entries"`
	LocalAddr   string    `json:"local_addr"`
	PeerAddr    string    `json:"peer_addr"`
	RequestLine string    `json:"request_line"`
	Timestamp   time.Time `json:"timestamp"`
}

type ResponseLogger struct {
	index int
}

const (
	LogIndexSkip = -1
)

var (
	responseLogs   []ConnectionLog
	responseLogsMu sync.Mutex
)

func NewResponseLogger(conn net.Conn) *ResponseLogger {
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

	return &ResponseLogger{index: index}
}

func (rl *ResponseLogger) setRequestLine(requestLine string) {
	responseLogsMu.Lock()
	defer responseLogsMu.Unlock()

	responseLogs[rl.index].RequestLine = requestLine

	log.Printf("%v >>> %v Received request: %s",
		responseLogs[rl.index].LocalAddr,
		responseLogs[rl.index].PeerAddr,
		responseLogs[rl.index].RequestLine)
}

func (rl *ResponseLogger) LogResponse(data string) {
	if rl.index == LogIndexSkip {
		return
	}

	for _, msg := range rl.formatResponseData(data) {
		rl.record(msg)
	}
}

func (rl *ResponseLogger) LogError(err error, bytesWritten int, content string) {
	if rl.index == LogIndexSkip {
		return
	}

	if err != nil {
		responseLogsMu.Lock()
		if responseLogs[rl.index].ConnError != "" {
			responseLogs[rl.index].ConnError = err.Error()
		}
		responseLogsMu.Unlock()
	}

	for _, msg := range rl.formatWriteError(err, bytesWritten, content) {
		rl.record(msg)
	}
}

func (rl *ResponseLogger) formatResponseData(data string) []string {
	var result []string
	maxContentLength := 200

	// Truncate the content if it's too long.
	if len(data) > maxContentLength {
		data = data[:maxContentLength] + "... (truncated)"
	}

	if strings.HasPrefix(data, "HTTP/") || strings.Contains(data, ": ") {
		result = append(result, data)
		return result
	}

	// Handle chunked data.
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

	// Handle empty data.
	if data == "" {
		result = append(result, "Empty-Response")
		return result
	}

	// Handle any other data (non-chunked body content).
	result = append(result, "Body: "+strings.TrimSpace(data))
	return result
}

func (rl *ResponseLogger) formatWriteError(err error, bytesWritten int, content string) []string {
	maxContentLength := 200
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "... (truncated)"
	}

	var result = []string{
		fmt.Sprintf("error writing %q after %d bytes: %v", content, bytesWritten, err),
	}

	return append(result, formatStackTrace(getStackFrames(6))...)
}

func (rl *ResponseLogger) record(format string, args ...interface{}) {
	if rl.index == LogIndexSkip {
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

func (rl *ResponseLogger) SetSkipLogging() {
	rl.index = LogIndexSkip
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
	w.logger.LogResponse(data)
}

func (w *responseWriter) logWriteError(err error, bytesWritten int, content string) {
	w.logger.LogError(err, bytesWritten, content)
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

	w.logResponseData(data)

	n, err := io.WriteString(w.conn, data)
	if err != nil {
		w.logWriteError(err, n, data)
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
	return w.write([]byte(fmt.Sprint(a...)))
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

	// Write the status line through writeToConnection
	w.writeToConnection(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode)))

	// Write all headers through writeToConnection
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

	// End headers with a blank line
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

// Error replies to the request with the specified error message and
// HTTP code. It does not otherwise end the request; the caller should
// ensure no further writes are done to w. The error message should be
// plain text.
func (w *responseWriter) httpError(error string, code int) {
	w.header().Set("Content-Type", "text/plain; charset=utf-8")
	w.header().Set("X-Content-Type-Options", "nosniff")
	w.writeHeader(code)
	w.print(error)
}

// handleConnection processes an incoming connection by setting a
// timeout, reading the request line, and routing it to the
// appropriate handler. It sets a connection timeout using
// setConnTimeout and logs any errors. Once the connection is
// processed or an error occurs, it ensures the connection is closed.
// The request is parsed for method, path, and protocol, and the
// request is routed based on the path.
func handleConnection(conn net.Conn) {
	logger := NewResponseLogger(conn)

	writer := responseWriter{
		conn:   conn,
		logger: logger,
	}

	clearTimeout, err := writer.setConnTimeout(30 * time.Second)
	if err != nil {
		//log.Printf("set connection timeout failed: %v", err)
		conn.Close()
		return
	}

	defer func() {
		logger.record("Connection closed")
		clearTimeout()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	requestLine, _, err := reader.ReadLine()
	if err != nil {
		log.Printf("%v: Error reading request line: %v\n", conn.LocalAddr(), err)
		return
	}

	logger.setRequestLine(string(requestLine))

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
	case "/access-logs":
		accessLogsHandler(&writer, method)
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
			// For HEAD requests, always set
			// Content-Length and never use chunked
			// encoding.
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		} else if useChunkedEncoding {
			w.header().Set("Transfer-Encoding", "chunked")
			for i := 0; i < additionalTEHeaders; i++ {
				w.header().Add("Transfer-Encoding", "chunked")
			}
		} else {
			w.header().Set("Content-Length", strconv.Itoa(len(body)))
		}
		w.writeHeader(statusCode)
		if method == http.MethodGet {
			if useChunkedEncoding {
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

func accessLogsHandler(w *responseWriter, method string) {
	// Skip logging for this request to avoid recursive logging.
	w.logger.SetSkipLogging()

	responseLogsMu.Lock()
	logs := make([]ConnectionLog, len(responseLogs))
	copy(logs, responseLogs)
	responseLogsMu.Unlock()

	// Marshal the logs to JSON
	logsJSON, err := json.Marshal(logs)
	if err != nil {
		w.httpError("Internal Server Error", http.StatusInternalServerError)
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
