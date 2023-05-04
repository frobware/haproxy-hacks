#!/usr/bin/bash

set -eu

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

tmpdir=$(mktemp -d)
trap 'rm -fr $tmpdir' EXIT
go run certgen/certgen.go > "$tmpdir/certs.sh"
DEST_CACRT="$(oc exec -n openshift-ingress -c router "$pod" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt)"
source "$tmpdir/certs.sh"
oc process DEST_CACRT="$DEST_CACRT" TLS_KEY="$TLS_KEY" TLS_CRT="$TLS_CRT" TLS_CACRT="$TLS_CACRT" -f ./destca-routes.yaml -o yaml | oc apply -f -
oc get routes
