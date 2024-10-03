// Package main implements a custom HTTP client designed to handle
// non-standard HTTP responses, including those with duplicate
// headers.
//
// Key features:
//
//   - Protocol Support: Handles both HTTP and HTTPS protocols.
//
//   - Redirect Handling: Supports up to 10 redirects, managing both
//     relative and absolute redirect URLs.
//
//   - Command-line Interface:
//
//   - -I flag to display only headers (similar to 'curl -I')
//
//   - -timeout flag for setting a custom timeout (default 5s)
//
//   - -k flag for ignoring certificate validation (similar to 'curl -k').
//
//   - Manual Response Parsing: Implements custom parsing of HTTP
//     responses, including status lines and headers.
//
//   - TLS Configuration: Uses InsecureSkipVerify when the '-k' flag
//     is set, bypassing certificate verification. This is suitable
//     for testing but not recommended for production use without
//     careful consideration.
//
//   - Flexible Output: Can display either full response body or just
//     headers based on user input.
//
// Usage:
//
//	go run client.go [-I] [-timeout duration] [-k] <URL>
//
// The program sends a GET request to the provided URL, parses the
// response, and prints the status, headers, and optionally the body.
// It supports both HTTP and HTTPS protocols.
package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const maxRedirects = 10

// readCloser wraps a reader and a closer to implement io.ReadCloser.
// This struct is used to combine separate io.Reader and io.Closer
// implementations into a single io.ReadCloser, which is required for
// HTTP response bodies.
//
// In the context of our custom HTTP client:
// - The 'reader' is a bufio.Reader reading from a net.Conn
// - The 'closer' is the net.Conn itself
//
// This separation allows us to use buffered reading for efficiency
// while still being able to close the underlying connection when
// needed.
type readCloser struct {
	reader io.Reader
	closer io.Closer
}

// Read implements the io.Reader interface.
func (rc *readCloser) Read(p []byte) (int, error) {
	return rc.reader.Read(p)
}

// Close implements the io.Closer interface.
func (rc *readCloser) Close() error {
	return rc.closer.Close()
}

// Ensure readCloser implements io.ReadCloser.
var _ io.ReadCloser = (*readCloser)(nil)

// httpGetWithoutSanitisation performs an HTTP GET request and returns
// the response without sanitising headers or body. This allows for
// handling of non-standard or non-compliant HTTP responses.
//
// Unlike the standard Go HTTP client, this function does not sanitise
// response headers or body, enabling inspection of responses from
// non-compliant servers or unusual HTTP implementations.
func httpGetWithoutSanitisation(ctx context.Context, rawURL string, insecure bool) (*http.Response, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("URL must include a scheme (http or https)")
	}

	redirectCount := 0

	for {
		host := parsedURL.Host
		if !strings.Contains(host, ":") {
			if parsedURL.Scheme == "https" {
				host += ":443"
			} else {
				host += ":80"
			}
		}
		path := parsedURL.RequestURI()
		if path == "" {
			path = "/"
		}
		useTLS := parsedURL.Scheme == "https"

		dialer := &net.Dialer{}

		var conn net.Conn
		if useTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: insecure,
			}
			conn, err = tls.DialWithDialer(dialer, "tcp", host, tlsConfig)
		} else {
			conn, err = dialer.DialContext(ctx, "tcp", host)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}

		if deadline, ok := ctx.Deadline(); ok {
			conn.SetDeadline(deadline)
		}

		reader := bufio.NewReader(conn)
		requestLine := fmt.Sprintf("GET %s HTTP/1.1\r\n", path)
		headers := fmt.Sprintf("Host: %s\r\nConnection: close\r\n\r\n", parsedURL.Hostname())
		_, err = conn.Write([]byte(requestLine + headers))
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("error writing request: %w", err)
		}

		statusLine, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("error reading status line: %w", err)
		}

		statusParts := strings.SplitN(strings.TrimSpace(statusLine), " ", 3)
		if len(statusParts) < 3 {
			conn.Close()
			return nil, fmt.Errorf("malformed status line: %v", statusLine)
		}

		headersMap := make(http.Header)
		for {
			line, isPrefix, err := reader.ReadLine()
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("error reading headers: %w", err)
			}
			headerLine := string(line)
			for isPrefix {
				line, isPrefix, err = reader.ReadLine()
				if err != nil {
					conn.Close()
					return nil, fmt.Errorf("error reading long header: %w", err)
				}
				headerLine += string(line)
			}

			if headerLine == "" {
				break
			}

			parts := strings.SplitN(headerLine, ": ", 2)
			if len(parts) == 2 {
				headersMap.Add(parts[0], parts[1])
			}
		}

		response := &http.Response{
			Status:     statusParts[1] + " " + statusParts[2],
			StatusCode: parseStatusCode(statusParts[1]),
			Proto:      statusParts[0],
			Header:     headersMap,
			Body: &readCloser{
				reader: reader,
				closer: conn,
			},
			Request: &http.Request{Method: "GET", URL: parsedURL},
		}

		if response.StatusCode >= http.StatusMultipleChoices && response.StatusCode < http.StatusBadRequest {
			redirectCount++
			if redirectCount > maxRedirects {
				response.Body.Close()
				return nil, fmt.Errorf("too many redirects")
			}

			location := response.Header.Get("Location")
			if location == "" {
				response.Body.Close()
				return nil, fmt.Errorf("redirect location not found")
			}

			// Handle relative and absolute URLs in the
			// "Location" header.
			redirectURL, err := parsedURL.Parse(location)
			if err != nil {
				response.Body.Close()
				return nil, fmt.Errorf("failed to parse redirect URL: %w", err)
			}

			// Validate that the redirect URL has a valid
			// hostname.
			if redirectURL.Host == "" {
				response.Body.Close()
				return nil, fmt.Errorf("redirect URL has empty host: %s", redirectURL.String())
			}

			parsedURL = redirectURL

			// Close the current response body and start
			// over for the redirect.
			response.Body.Close()
			continue
		}

		// Return the final response if it's not a redirect.
		return response, nil
	}
}

// parseStatusCode converts a string HTTP status code to an integer.
// It returns 0 if the conversion fails.
func parseStatusCode(status string) int {
	code, err := strconv.Atoi(status)
	if err != nil {
		return 0 // Invalid status code
	}
	return code
}

// readRawBody reads and returns the raw body of the HTTP response.
//
// This function reads the entire content of the provided io.Reader
// without any processing or decoding. It is particularly useful for
// debugging purposes or when the raw, unprocessed response body is
// required.
//
// The function uses io.ReadAll internally, which means it will
// continue reading until an error occurs or EOF is reached. This
// approach is suitable for most scenarios but may not be ideal for
// very large responses or in memory-constrained environments.
//
// Parameters:
//   - reader: An io.Reader from which to read the raw body.
//
// Returns:
//   - []byte: The raw body content.
//   - error: An error if any issues occur during reading. This will
//     typically be nil unless there's an I/O error.
//
// Note: This function reads the entire body into memory.
func readRawBody(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}

// readChunkedBody reads and decodes a chunked HTTP response body.
//
// This function implements a lenient chunked transfer decoding as
// specified in RFC 7230. It handles premature EOFs and incomplete
// chunks, which may occur with non-compliant servers.
//
// The function reads chunks until it encounters a chunk of size 0 or
// reaches EOF. For each chunk, it reads the size, converts it from
// hexadecimal to integer, reads the chunk data, and appends it to the
// body.
//
// If an EOF is encountered at any point (reading chunk size, chunk
// data, or chunk ending), the function returns the data read so far
// without an error. This behaviour allows handling of incomplete or
// non-compliant chunked responses.
//
// Parameters:
//   - reader: An io.Reader from which to read the chunked body.
//
// Returns:
//   - []byte: The decoded body content.
//   - error: An error if any issues occur during reading or parsing,
//     except for EOF.
//
// Possible errors:
//   - "error reading chunk size": Failed to read the chunk size line.
//   - "error parsing chunk size": Failed to parse the chunk size as a
//     hexadecimal integer.
//   - "error reading chunk": Failed to read the chunk data (excluding
//     EOF).
//   - "error reading chunk ending": Failed to read the CRLF at the end
//     of a chunk (excluding EOF).
//
// Note: This function reads the entire body into memory.
func readChunkedBody(reader io.Reader) ([]byte, error) {
	var body []byte
	bufReader := bufio.NewReader(reader)

	for {
		// Read the chunk size.
		sizeStr, err := bufReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// If we've reached EOF, return what we've read so far
				return body, nil
			}
			return nil, fmt.Errorf("error reading chunk size: %w", err)
		}
		sizeStr = strings.TrimSpace(sizeStr)

		// Convert chunk size from hex to int
		size, err := strconv.ParseInt(sizeStr, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing chunk size: %w", err)
		}

		// If chunk size is 0, we've reached the end
		if size == 0 {
			break
		}

		// Read the chunk
		chunk := make([]byte, size)
		n, err := io.ReadFull(bufReader, chunk)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// If we've reached EOF, append what we've read and return
				body = append(body, chunk[:n]...)
				return body, nil
			}
			return nil, fmt.Errorf("error reading chunk: %w", err)
		}
		body = append(body, chunk...)

		// Read and discard the CRLF at the end of the chunk
		_, err = bufReader.Discard(2)
		if err != nil {
			if err == io.EOF {
				// If we've reached EOF, return what we've read so far
				return body, nil
			}
			return nil, fmt.Errorf("error reading chunk ending: %w", err)
		}
	}
	return body, nil
}

func main() {
	var (
		insecure    bool
		showHeaders bool
		showRawBody bool
		timeoutStr  string
	)

	flag.BoolVar(&insecure, "k", false, "Allow insecure server connections when using TLS (similar to curl -k)")
	flag.BoolVar(&showHeaders, "I", false, "Show headers only (similar to curl -I)")
	flag.BoolVar(&showRawBody, "raw", false, "Show raw body without processing chunked encoding")
	flag.StringVar(&timeoutStr, "timeout", "5s", "Set timeout duration (e.g., 5s, 500ms)")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Printf("Usage: %s [options] <URL>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	rawURL := args[0]

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		fmt.Printf("Invalid timeout duration: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := httpGetWithoutSanitisation(ctx, rawURL, insecure)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}

	defer response.Body.Close()

	if showHeaders {
		fmt.Printf("%s %s\n", response.Proto, response.Status)
		for key, values := range response.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", strings.ToLower(key), value)
			}
		}
	}

	if !showHeaders || showRawBody {
		var bodyBag []byte
		var err error

		if showRawBody {
			bodyBag, err = readRawBody(response.Body)
		} else if strings.EqualFold(response.Header.Get("Transfer-Encoding"), "chunked") {
			bodyBag, err = readChunkedBody(response.Body)
		} else {
			bodyBag, err = io.ReadAll(response.Body)
		}

		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
		} else {
			fmt.Printf("%s", string(bodyBag))
		}
	}
}
