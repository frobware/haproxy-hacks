#!/usr/bin/env bash

# Script to start port forwarding for Prometheus in OpenShift.

prometheus_namespace="openshift-monitoring"

# Step 1: Ensure you are logged into the OpenShift cluster using `oc login`.
# Step 2: Dynamically determine the Prometheus pod name and set up port forwarding.

prometheus_pod=$(oc get pods -n $prometheus_namespace -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -z "$prometheus_pod" ]; then
    echo "Prometheus pod not found in namespace $prometheus_namespace with label 'app.kubernetes.io/name=prometheus'."
    echo "Please check the available pods and their labels using: oc get pods -n $prometheus_namespace --show-labels"
    exit 1
fi

# Step 3: Set up port forwarding to access Prometheus on
# localhost:9090.

echo "Setting up port-forward to Prometheus pod: $prometheus_pod"
oc port-forward -n $prometheus_namespace $prometheus_pod 9090:9090 &

port_forward_pid=$!
echo "Port forwarding pid: $port_forward_pid"
echo $port_forward_pid > /tmp/prometheus_port_forward_pid

echo "Port forwarding is now active. You can query Prometheus using 'query_prometheus.sh'."
