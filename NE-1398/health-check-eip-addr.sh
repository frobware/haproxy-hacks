#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "usage: $0 <ip-address>" >&2
    exit 1
fi

curl -I http://"$1"/_______internal_router_healthz
