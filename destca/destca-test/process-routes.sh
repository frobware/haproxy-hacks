#!/usr/bin/env bash

set -eu

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
route_spec=$2

tls_key="$(cat $certdir/key.pem)"
tls_crt="$(cat $certdir/fullchain.pem)"
tls_cacrt="$(cat $certdir/ca.cer)"

dest_cacrt="$(oc exec -n openshift-ingress -c router "$pod" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt)"
oc process DEST_CACRT="$dest_cacrt" TLS_KEY="$tls_key" TLS_CRT="$tls_crt" TLS_CACRT="$tls_cacrt" -f $route_spec -o yaml | oc delete --ignore-not-found -f -
oc process DEST_CACRT="$dest_cacrt" TLS_KEY="$tls_key" TLS_CRT="$tls_crt" TLS_CACRT="$tls_cacrt" -f $route_spec -o yaml | oc apply -f -
