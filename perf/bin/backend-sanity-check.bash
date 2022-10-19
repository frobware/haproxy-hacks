#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"

for name in $(backend_names_sorted); do
    url="https://$name.${domain}:${BACKEND_PORTS[$name]}/1024.html"
    curl --connect-timeout 1 -4 -o /dev/null -k -L -s -w "GET $url status %{http_code}\n" "${url}"
done
