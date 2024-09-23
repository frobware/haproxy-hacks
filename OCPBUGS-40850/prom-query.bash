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

translate_curl_error() {
    local curl_error=$1
    case $curl_error in
        6)  echo "Curl error: Could not resolve host." >&2 ;;
        7)  echo "Curl error: Failed to connect to host." >&2 ;;
        28) echo "Curl error: Operation timed out." >&2 ;;
        35) echo "Curl error: SSL handshake failed." >&2 ;;
        52) echo "Curl error: Empty response from server." >&2 ;;
        56) echo "Curl error: Failure in receiving network data." >&2 ;;
        *)  echo "Curl error: Unknown error (code: $curl_error)." >&2 ;;
    esac
}

query_prometheus() {
    local query_file=$1
    local duration=$2

    if [[ ! -f $query_file ]]; then
        echo "Query file not found: $query_file" >&2
        return 1
    fi

    local query
    query=$(grep -v '^#' "$query_file" | tr '\n' ' ')

    if [[ -z $query ]]; then
        echo "Query is empty after processing: $query_file" >&2
        return 1
    fi

    # Replace placeholder in the query
    query="${query//$duration_placeholder/$duration}"

    # Execute curl and capture both status code and response
    local http_status
    local curl_status
    http_status=$(curl -s -w "%{http_code}" -G 'http://localhost:9090/api/v1/query' --data-urlencode "query=$query" -o /tmp/prom_response)
    curl_status=$?

    # Check if curl command failed
    if [[ $curl_status -ne 0 ]]; then
        translate_curl_error $curl_status
        return 1
    fi

    # Check if HTTP status code is not 200 (OK)
    if [[ $http_status -ne 200 ]]; then
        echo "HTTP request failed with status code: $http_status" >&2
        return 1
    fi

    # Read the actual response content from temporary file
    local response
    response=$(cat /tmp/prom_response)

    # Check if the response is valid JSON
    if ! echo "$response" | jq empty >/dev/null 2>&1; then
        echo "Invalid JSON response from Prometheus." >&2
        return 1
    fi

    # Clean up temporary file
    rm -f /tmp/prom_response

    # Return the valid JSON response
    echo "$response"
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
