#!/usr/bin/env bash

# Script to start port forwarding for Prometheus in OpenShift.

prometheus_namespace="openshift-monitoring"
pid_file="/tmp/prometheus_port_forward_pid"
foreground=false  # By default, run in detached mode (background)

while getopts "f" opt; do
    case $opt in
        f)
            foreground=true  # If -f is passed, run in foreground
            ;;
        *)
            echo "Usage: $0 [-f]"
            exit 1
            ;;
    esac
done

# Preliminary check for an existing port-forward process
if [ -f "$pid_file" ]; then
    existing_pid=$(cat "$pid_file")
    if kill -0 "$existing_pid" 2>/dev/null; then
        echo "An existing port-forward process is already running with PID: $existing_pid"
        echo "Check logs or stop the existing process if necessary."
        exit 0
    else
        echo "Found a stale PID file with PID: $existing_pid. Cleaning up..."
        rm -f "$pid_file"
    fi
fi

# Step 1: Ensure you are logged into the OpenShift cluster using `oc login`.
# Step 2: Dynamically determine the Prometheus pod name and set up port forwarding.

if ! prometheus_pod=$(oc get pods -n "$prometheus_namespace" -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); then
    echo "Error getting Prometheus pod name." >&2
    exit 1
fi

if [ -z "$prometheus_pod" ]; then
    echo "Prometheus pod not found in namespace $prometheus_namespace with label 'app.kubernetes.io/name=prometheus'."
    echo "Please check the available pods and their labels using: oc get pods -n $prometheus_namespace --show-labels"
    exit 1
fi

# Step 3: Set up port forwarding to access Prometheus on localhost:9090.

echo "Setting up port-forward to Prometheus pod: $prometheus_pod"

if $foreground; then
    oc port-forward -n "$prometheus_namespace" "$prometheus_pod" 9090:9090
else
    log_file=$(mktemp /tmp/prometheus_port_forward_XXXXXX.log)
    echo "Logging output to $log_file"

    oc port-forward -n "$prometheus_namespace" "$prometheus_pod" 9090:9090 >"$log_file" 2>&1 &

    port_forward_pid=$!
    echo "Port forwarding pid: $port_forward_pid"
    echo $port_forward_pid > "$pid_file"

    max_attempts=10
    attempt=0
    check_interval=2

    while [ $attempt -lt $max_attempts ]; do
        if ! kill -0 $port_forward_pid 2>/dev/null; then
            echo "Port forwarding process has terminated unexpectedly. Check the log file: $log_file" >&2
            rm -f "$pid_file"
            exit 1
        fi

        http_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/-/ready)

        if [ "$http_status" -eq 200 ]; then
            echo "Port forwarding is now active and Prometheus is ready. PID: $port_forward_pid"
            echo "Check logs: $log_file"
            exit 0
        else
            echo "Checking if Prometheus is ready (attempt $((attempt+1))/$max_attempts)..."
        fi

        sleep $check_interval
        attempt=$((attempt+1))
    done

    echo "Prometheus did not become ready after $max_attempts attempts. Check the log file: $log_file" >&2
    rm -f "$pid_file"
    exit 1
fi
