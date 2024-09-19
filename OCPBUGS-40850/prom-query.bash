#!/usr/bin/env bash

# Script to run Prometheus queries with optional post-processing using
# jq. A query file must be provided. The script substitutes a
# placeholder for duration in the query. If a corresponding .jq file
# exists, it will apply jq post-processing unless the raw option is
# used.

set -eu

duration_placeholder="{{DURATION}}"

usage() {
    echo "Usage: ${0##*/} [options] <query-file> [duration]"
    echo ""
    echo "Options:"
    echo "  -r, --raw    Output raw Prometheus query result (skip jq post-processing)"
    echo ""
    echo "Arguments:"
    echo "  <query-file>  The path to the query file (with .query) to run."
    echo "  [duration]    Optional duration to replace in the query (default: '1h')."
    echo ""
    echo "Examples:"
    echo "  ${0##*/} /path/to/query-file.query 5m"
    echo "  ${0##*/} --raw /path/to/query-file.query"
}

query_prometheus() {
    local query_file=$1
    local duration=$2

    local query
    query=$(grep -v '^#' "$query_file" | tr '\n' ' ')
    query="${query//$duration_placeholder/$duration}"

    curl -s -G 'http://localhost:9090/api/v1/query' --data-urlencode "query=$query"
}

jq_post_process=true
duration="1h"

while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--raw)
            jq_post_process=false
            shift
            ;;
        -*)
            usage
            exit 1
            ;;
        *)
            query_file=$1
            duration="${2:-$duration}"
            break
            ;;
    esac
done

if [[ -z ${query_file:-} ]]; then
    usage
    exit 1
fi

if [[ ! -f "$query_file" ]]; then
    echo "Error: Query file '$query_file' does not exist."
    exit 1
fi

response=$(query_prometheus "$query_file" "$duration")
jq_file="${query_file%.query}.jq"
if $jq_post_process && [[ -f "$jq_file" ]]; then
    echo "$response" | jq -r -f "$jq_file"
else
    echo "$response"
fi
