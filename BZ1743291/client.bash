#!/bin/bash

echo "Spawning ${1:-1} websocket connections"
data=$(perl -E 'say "=" x 32768')

for i in $(seq 1 ${1:-1}); do
    yes | websocat -n -t ws://127.0.0.1:4242/echo stdio:
done
