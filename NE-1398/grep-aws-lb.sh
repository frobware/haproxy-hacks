#!/usr/bin/env bash

# Enumerate all pods in all namespaces and grep their logs for
# 'aws-load-balancer'.

K="kubectl"

if ! command -v $K &> /dev/null; then
    K="oc"
    if ! command -v $K &> /dev/null; then
        echo "Neither kubectl nor oc command is available. Exiting."
        exit 1
    fi
fi

namespaces=$($K get namespaces -o jsonpath="{.items[*].metadata.name}")

echo "Searching for 'aws-load-balancer' in logs of all pods across all namespaces..."

for ns in $namespaces; do
    pods=$($K get pods -n "$ns" -o jsonpath="{.items[*].metadata.name}")
    for pod in $pods; do
        echo "Checking logs of pod $pod in namespace $ns for 'aws-load-balancer'..."
        $K logs "$pod" -n "$ns" | grep "aws-load-balancer" && echo "Found in $pod in namespace $ns"
    done
done
