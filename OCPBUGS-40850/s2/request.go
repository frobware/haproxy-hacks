// request.go

package main

import (
	"bufio"
	"net"
	"net/http"
)

// CustomRequestReader parses the HTTP request from the connection.
func CustomRequestReader(conn net.Conn) (*http.Request, error) {
	// Create a buffered reader for the connection
	reader := bufio.NewReader(conn)

	// Parse the HTTP request using the standard library's readRequest function
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}

	// Manually set the RemoteAddr for logging purposes.
	req.RemoteAddr = conn.RemoteAddr().String()

	return req, nil
}
