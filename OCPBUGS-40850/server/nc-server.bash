#!/usr/bin/env bash

trap 'exit 0' EXIT INT

while true; do
    {
        printf "HTTP/1.1 200 OK\r\n"
        printf "Date: %s\r\n" "$(date)"
        printf "Content-Type: text/plain; charset=utf-8\r\n"
        printf "Transfer-Encoding: chunked\r\n"
        printf "Transfer-Encoding: chunked, gzip\r\n"
        printf "Foo: Bar\r\n"
        printf "Set-Cookie: testcookie=value; path=/\r\n"
        printf "Foo: Baz\r\n"
        printf "Foo: Baz\r\n"
        printf "\r\n"     # Blank line to end the headers
        printf "4\r\n"    # Chunk size (in hexadecimal)
        printf "Test\r\n" # Chunk data
        printf "A\r\n"
        printf "HelloWorld\r\n"
        printf "1\r\n"
        printf "\n\r\n"
        printf "0\r\n"    # Last chunk to signal the end of the chunked transfer
        printf "\r\n"     # End of the response
    } | nc -l ${1:-8080}
done
