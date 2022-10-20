#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"

for name in $(backend_names_sorted); do
    echo "$name ${BACKEND_HTTPS_PORTS[$name]}"
done | go run "${thisdir}/gen-mb-requests.go" -domain $domain "$@"
