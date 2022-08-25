#! /usr/bin/env bash

set -eu
set -o pipefail

: "${NAMESPACE:=default}"

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

if [[ $? -ne 0 ]]; then
    echo "get domain failed" >&2
    exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "$tmpdir"' EXIT

oc process -o yaml -p NAMESPACE="$NAMESPACE" -f deployment.yaml
