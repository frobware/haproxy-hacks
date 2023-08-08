#!/usr/bin/env bash

set -eux

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

if [[ $# -lt 2 ]]; then
    echo "Specify a certdir and a route.yaml"
    exit 1
fi

certdir=$1
route_file=$2
dest_cacrt=$3

if [[ -z "$dest_cacrt" ]]; then
    tls_key="$(cat "$certdir/tls.key")"
    tls_crt="$(cat "$certdir/tls.crt")"
else
    tls_key=""
    tls_crt=""
fi

oc process DEST_CACRT="$dest_cacrt" TLS_KEY="$tls_key" TLS_CRT="$tls_crt" -f "$route_file" -o yaml | oc delete --ignore-not-found -f -
oc process DEST_CACRT="$dest_cacrt" TLS_KEY="$tls_key" TLS_CRT="$tls_crt" -f "$route_file" -o yaml | oc apply -f -
