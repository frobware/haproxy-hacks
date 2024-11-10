#!/usr/bin/env bash

set -eu

# This script is designed to reproduce a potential issue in HAProxy
# connection handling when switching between backends under frequent
# reloads. Specifically, it aims to observe whether HAProxy maintains
# stale connections to an old backend after configuration updates and
# a reload, a scenario that could occur in environments like OpenShift
# where frequent backend switches or reloads happen.
#
# Summary of the steps:
#
# 1. Initial Setup:
#    - Two backend containers (`backend1` and `backend2`) are started,
#      each running an instance of an HTTP server (from a predefined
#      `nginx-alpine` image).
#    - HAProxy is configured to initially route requests to `backend1`
#      and is started in daemon mode.
#
# 2. Backend Switching Logic:
#    - The script runs an infinite loop that alternates between
#      `backend1` and `backend2` by modifying the HAProxy configuration
#      (`haproxy.config`) and reloading HAProxy with the new backend
#      setting.
#    - After each switch, the script sends an HTTP request to HAProxy on
#      the frontend port (`$FRONTEND_PORT`), capturing the response to
#      determine which backend served the request.
#
# 3. Verification and Delay:
#    - After each request, the response is checked to ensure it matches
#      the expected backend (`backend1` or `backend2`).
#    - The `REQUEST_DELAY` variable specifies a wait time between
#      backend switches to simulate real-world delays that may be present
#      in production environments.
#    - If a request unexpectedly reaches an old backend (e.g., `backend1`
#      after switching to `backend2`), the script will print an error and
#      exit, indicating that the issue has reproduced.
#
# 4. Cleanup:
#    - The `cleanup` function ensures that resources are freed when the
#      script exits, stopping and removing the backend containers and
#      deleting temporary files.

: "${BACKEND1_PORT:=18081}"
: "${BACKEND2_PORT:=18082}"
: "${FRONTEND_PORT:=18080}"
: "${HAPROXY_BIN:=haproxy}"
: "${PODMAN:=podman}"
: "${REQUEST_DELAY:=${1:-0}}"

temp_dir=$(mktemp -d)
pid_file="$temp_dir/haproxy.pid"
cleanup_called=false

function cleanup {
    if [ "$cleanup_called" = true ]; then
        return
    fi
    cleanup_called=true
    echo "Cleaning up"
    podman stop backend1 backend2 > /dev/null 2>&1 || true
    podman rm backend1 backend2 > /dev/null 2>&1 || true
    if [ -f "$pid_file" ]; then
        kill "$(cat "$pid_file")" || true
    fi
    rm -rf "$temp_dir"
}

trap cleanup EXIT INT TERM

function start_container {
    local name=$1
    local port=$2
    local image=$3

    if [ "$(podman ps -q -f name='$name')" ]; then
        echo "Container $name is already running."
    else
        if [ "$(podman ps -a -q -f name='$name')" ]; then
            echo "Container $name exists but is stopped. Starting it."
            podman start "$name"
        else
            echo "Creating and starting container $name on port $port."
            podman run -d --name "$name" -p "$port":8080 "$image"
        fi
    fi
}

start_container backend1 "$BACKEND1_PORT" quay.io/openshifttest/nginx-alpine@sha256:04f316442d48ba60e3ea0b5a67eb89b0b667abf1c198a3d0056ca748736336a0
start_container backend2 "$BACKEND2_PORT" quay.io/openshifttest/nginx-alpine@sha256:04f316442d48ba60e3ea0b5a67eb89b0b667abf1c198a3d0056ca748736336a0

backend1_id=$(podman ps -qf "name=backend1")
backend2_id=$(podman ps -qf "name=backend2")

echo "backend1 container ID: $backend1_id on port $BACKEND1_PORT"
echo "backend2 container ID: $backend2_id on port $BACKEND2_PORT"

cat > "${temp_dir}/haproxy.config" <<EOF
global
    daemon
    log /dev/log local0 info

defaults
    mode http
    option httplog
    option dontlognull
    option log-health-checks
    timeout connect 5000
    timeout client 50000
    timeout server 50000
    timeout client-fin 1s
    timeout server-fin 1s
    timeout http-request 10s
    timeout http-keep-alive 300s
    option idle-close-on-response

frontend fe_test
    log global
    bind *:$FRONTEND_PORT
    default_backend be_service

backend be_service
    server backend1 127.0.0.1:$BACKEND1_PORT check
EOF

"$HAPROXY_BIN" -v
"$HAPROXY_BIN" -f "${temp_dir}/haproxy.config" -p "$pid_file"

function test_backend {
    echo "Making request to frontend on port $FRONTEND_PORT"

    if ! response=$(curl -s -i http://localhost:"$FRONTEND_PORT"); then
        echo "Error: Failed to connect to frontend on port $FRONTEND_PORT" >&2
        exit 1
    fi

    echo "Full response:"
    echo "$response"

    echo "$response" | awk '/Hello-OpenShift/ {print $2}'
}

function check_response {
    local response="$1"
    local expected_backend_id="$2"
    if [[ $response == *"$expected_backend_id"* ]]; then
        echo "Response from request: $response"
    else
        echo "Error: Unexpected response $response; expected $expected_backend_id" >&2
        exit 1
    fi
}

function switch_backend {
    local backend_port=$1
    echo "Switching backend to port $backend_port"
    sed -i "s/127.0.0.1:[0-9]*/127.0.0.1:$backend_port/" "${temp_dir}/haproxy.config"
    "$HAPROXY_BIN" -f "${temp_dir}/haproxy.config" -p "$pid_file" -sf "$(cat "$pid_file")"
}

function test_and_verify_backend {
    local expected_backend_id=$1

    if [[ "$REQUEST_DELAY" -gt 0 ]]; then
        echo "Waiting for $REQUEST_DELAY seconds before the next request"
        sleep "$REQUEST_DELAY"
    fi

    new_response=$(test_backend)
    check_response "$new_response" "$expected_backend_id"
}

# Initial request to verify backend1 responds.
echo "Initial request to backend1"
initial_response=$(test_backend)
check_response "$initial_response" "$backend1_id"

# Infinite loop to alternate between backends.
while true; do
    switch_backend "$BACKEND2_PORT"
    test_and_verify_backend "$backend2_id"

    switch_backend "$BACKEND1_PORT"
    test_and_verify_backend "$backend1_id"
done
