#!/bin/bash

socat -x TCP-LISTEN:1025,reuseaddr,fork SYSTEM:'
    {
        printf "HTTP/1.1 200 OK\r\n"
        printf "Date: %s\r\n" "$(date -u)"
        printf "Content-Type: text/plain; charset=utf-8\r\n"
        printf "Foo: Bar\r\n"
        printf "Transfer-Encoding: chunked\r\n"
        printf "Foo: Baz\r\n"
        printf "Set-Cookie: testcookie=value; path=/\r\n"
        printf "\r\n"
        printf "4\r\n"
        printf "Test\r\n"
        printf "A\r\n"
        printf "HelloWorld\r\n"
        printf "1\r\n"
        printf "\n\r\n"
        printf "0\r\n"
        printf "\r\n"
    }',pipes

