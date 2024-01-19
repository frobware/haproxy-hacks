#!/usr/bin/env bash

current_namespace="$(oc project -q)"

if [[ -z "$current_namespace" ]]; then
    echo "Please log in to OpenShift and set an active project."
    exit 1
fi

routes=$(oc get routes -o jsonpath='{.items[*].metadata.name}' -n "$current_namespace")
route_hostnames=$(oc get routes -o jsonpath='{.items[*].spec.host}' -n "$current_namespace")

if [[ -z "$routes" ]]; then
    echo "No routes found in namespace ($current_namespace)."
    exit 1
fi

for host in $route_hostnames; do
    result=$(curl -s -k -L --http2-prior-knowledge "$host")
    echo "proto result: $result for $host"
done
