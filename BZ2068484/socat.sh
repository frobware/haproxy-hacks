#!/usr/bin/env bash

socat -T 1 -d -d tcp-l:${1:-1936},reuseaddr,reuseport,fork,crlf system:"echo -e \"\\\"HTTP/1.0 200 OK\\\nDocumentType: text/html\\\nHeader: Set-Cookie "X=Y";\\\n\\\n<html>date: \$\(date\)<br>server:\$SOCAT_SOCKADDR:\$SOCAT_SOCKPORT<br>client: \$SOCAT_PEERADDR:\$SOCAT_PEERPORT\\\n<pre>\\\"\"; cat; echo -e \"\\\"\\\n</pre></html>\\\"\""
