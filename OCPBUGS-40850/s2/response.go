// response.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
)

type ResponseFragment struct {
	Data []byte
	Type string // e.g., "status_line", "header", "body"
}

type CustomResponseWriter struct {
	conn       net.Conn
	header     http.Header
	status     int
	body       []byte
	useChunked bool
	trailers   http.Header
	request    *http.Request
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
	var formattedFrames []string

	for _, frame := range frames {
		file := filepath.Base(frame.File)
		formattedFrames = append(formattedFrames, fmt.Sprintf("%s (%s:%d)", frame.Function, file, frame.Line))
	}

	return formattedFrames
}

// NewCustomResponseWriter creates a new instance of a custom response writer.
func NewCustomResponseWriter(conn net.Conn, req *http.Request) *CustomResponseWriter {
	return &CustomResponseWriter{
		conn:     conn,
		header:   make(http.Header),
		trailers: make(http.Header),
		request:  req, // Store the request to check if trailers are allowed
	}
}

// Header returns the headers map for the response.
func (crw *CustomResponseWriter) Header() http.Header {
	return crw.header
}

// WriteHeader stores the status code for the response.
func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.status = statusCode
}

// Write accumulates the body content of the response.
func (crw *CustomResponseWriter) Write(data []byte) (int, error) {
	crw.body = append(crw.body, data...)
	return len(data), nil
}

// EnableChunked sets chunked transfer encoding for the response.
func (crw *CustomResponseWriter) EnableChunked() {
	crw.useChunked = true
	crw.Header().Set("Transfer-Encoding", "chunked")
}

// AddTrailer adds a trailer header to the response.
func (crw *CustomResponseWriter) AddTrailer(key, value string) {
	crw.trailers.Add(key, value)
}

// shouldSendTrailers checks if the client supports trailers by looking for the "TE: trailers" header.
func (crw *CustomResponseWriter) shouldSendTrailers() bool {
	teHeader := crw.request.Header.Get("TE")
	return strings.Contains(teHeader, "trailers")
}

// Helper function to wrap write operations and log them.
func (crw *CustomResponseWriter) writeAndLog(content string) (int, error) {
	n, err := crw.conn.Write([]byte(content))
	crw.logWriteStep(content, n, err)
	return n, err
}

// Log each write step with content and error if any.
func (crw *CustomResponseWriter) logWriteStep(content string, bytesWritten int, err error) {
	if err != nil {
		log.Printf("Write error:")
		log.Printf("  Content: %q", content)
		log.Printf("  Bytes written: %d", bytesWritten)
		log.Printf("  Error: %v", err)
		log.Printf("Stack trace:")

		for i, frame := range formatStackTrace(getStackFrames(4)) {
			log.Printf("  " + frame)
			if i > 5 {
				break // some brevity
			}
		}
	} else {
		log.Printf("Write success: %s\n", content)
	}
}

func (crw *CustomResponseWriter) Finalise(localAddr net.Addr) error {
	// Reject methods other than GET and HEAD.
	if crw.request.Method != http.MethodGet && crw.request.Method != http.MethodHead {
		statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		_, err := crw.writeAndLog(statusLine)
		if err != nil {
			return err
		}
		_, err = crw.writeAndLog("Allow: GET, HEAD\r\n\r\n")
		return err
	}

	// Write status line
	if crw.status == 0 {
		crw.WriteHeader(http.StatusOK) // Default status code if none is set
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", crw.status, http.StatusText(crw.status))
	_, err := crw.writeAndLog(statusLine)
	if err != nil {
		return err
	}

	// Write headers
	for k, v := range crw.header {
		for _, val := range v {
			headerLine := fmt.Sprintf("%s: %s\r\n", k, val)
			_, err := crw.writeAndLog(headerLine)
			if err != nil {
				return err
			}
		}
	}

	// Handle trailers (only if chunked transfer encoding is used and client supports them)
	if len(crw.trailers) > 0 && crw.useChunked && crw.shouldSendTrailers() {
		trailerKeys := make([]string, 0, len(crw.trailers))
		for k := range crw.trailers {
			trailerKeys = append(trailerKeys, k)
		}
		trailerHeader := fmt.Sprintf("Trailer: %s\r\n", strings.Join(trailerKeys, ", "))
		_, err := crw.writeAndLog(trailerHeader)
		if err != nil {
			return err
		}
	}

	// End headers section
	_, err = crw.writeAndLog("\r\n")
	if err != nil {
		return err
	}

	// For HEAD requests, we stop here - no body should be sent.
	if crw.request.Method == http.MethodHead {
		return nil
	}

	// Write body or chunked data for non-HEAD requests
	if crw.useChunked {
		chunkSize := 8
		for i := 0; i < len(crw.body); i += chunkSize {
			end := i + chunkSize
			if end > len(crw.body) {
				end = len(crw.body)
			}
			chunk := crw.body[i:end]

			// Write chunk size
			chunkSizeLine := fmt.Sprintf("%x\r\n", len(chunk))
			_, err = crw.writeAndLog(chunkSizeLine)
			if err != nil {
				return err
			}

			// Write chunk data
			_, err = crw.writeAndLog(string(chunk))
			if err != nil {
				return err
			}

			// Write chunk terminator
			_, err = crw.writeAndLog("\r\n")
			if err != nil {
				return err
			}
		}

		// Write the entire body as a single chunk
		if false && len(crw.body) > 0 {
			chunkSize := len(crw.body)
			chunkSizeLine := fmt.Sprintf("%x\r\n", chunkSize)
			_, err = crw.writeAndLog(chunkSizeLine)
			if err != nil {
				return err
			}
			_, err = crw.writeAndLog(string(crw.body))
			if err != nil {
				return err
			}
			_, err = crw.writeAndLog("\r\n")
			if err != nil {
				return err
			}
		}

		// End chunked transfer encoding
		_, err = crw.writeAndLog("0\r\n")
		if err != nil {
			return err
		}

		// Write trailers only if the client supports them
		if crw.shouldSendTrailers() {
			for k, v := range crw.trailers {
				for _, val := range v {
					trailerLine := fmt.Sprintf("%s: %s\r\n", k, val)
					_, err = crw.writeAndLog(trailerLine)
					if err != nil {
						return err
					}
				}
			}
		}

		// Always end with CRLF, regardless of trailers
		_, err = crw.writeAndLog("\r\n")
		if err != nil {
			return err
		}
	} else {
		// Write non-chunked body
		if len(crw.body) > 0 {
			_, err = crw.writeAndLog(string(crw.body))
		}
	}

	return err
}
