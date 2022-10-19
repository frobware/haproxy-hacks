#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"

for name in $(backend_names_sorted); do
    echo "$name.${domain} ${BACKEND_PORTS[$name]}"
done | go run gen-mb-requests.go "$@"
