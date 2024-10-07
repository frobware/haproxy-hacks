#!/usr/bin/env bash

# Function to normalize HTTP responses by removing dynamic headers and carriage returns
normalize_output() {
    sed '/^Date: /d' | sed '/^\*/d' | tr -d '\r'
}

echo -e "GET $1 HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n" | nc localhost "${2:-1051}" | normalize_output
