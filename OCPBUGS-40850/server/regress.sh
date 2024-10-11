#!/usr/bin/env bash

set -eu

usage() {
    echo "Usage: $0 <port> <output-dir>"
    echo "  <port>       - Server port to connect to (e.g., 1051)"
    echo "  <output_dir> - Directory to save GET request output"
    exit 1
}

if [ $# -ne 2 ]; then
    echo "Error: Missing arguments." >&2
    usage
fi

port="$1"
output_dir="$2"

mkdir -p "$output_dir"

if ! endpoints=$(curl -sS "http://localhost:${1:-1051}/discovery"); then
    echo "no endpoints" >&2
    exit 1
fi

if [ -z "$endpoints" ]; then
    echo "no endpoints" >&2
    exit 1
fi

for i in $endpoints; do
    ./get.sh "$i" "$port" > "$output_dir/$i.txt"
done
