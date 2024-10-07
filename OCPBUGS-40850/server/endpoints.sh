#!/usr/bin/env bash

curl -sS "http://localhost:${1:-1051}/" | grep -Eo '^/[a-z-]+' | sort | grep -v access-logs
