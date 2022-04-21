#!/usr/bin/env bash

set -euo pipefail

echo "show info" | socat ${1:-/tmp/haproxy.sock} stdio | grep -i -e '^maxsock:' -e '^maxconn:'
