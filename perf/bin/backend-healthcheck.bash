#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"

declare -a failures=()

for name in $(backend_names_sorted); do
    url="$name.${domain}:${BACKEND_PORTS[$name]}/1024.html"
    if ! curl --connect-timeout 3 -o /dev/null -q -k -L -s -w "GET $url status %{http_code}\n" "https://${url}"; then
	failures+=($url)
    fi
done

if [[ ${#failures[*]} -gt 0 ]]; then
    printf "%s\n" ${failures[@]} | sort -V
    echo "${#failures[*]} FAILURES"
    exit 1
fi

exit 0

