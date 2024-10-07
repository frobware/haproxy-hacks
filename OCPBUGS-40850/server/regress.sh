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

mkdir -p "tests/$output_dir"

for i in $(./endpoints.sh "$port"); do
    ./get.sh "$i" "$port" > "tests/$output_dir/$i.txt"
done
