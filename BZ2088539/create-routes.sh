#!/usr/bin/env bash

set -eu
set -o pipefail

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

if [[ $? -ne 0 ]]; then
    echo "get domain failed" >&2
    exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "$tmpdir"' EXIT

go run certgen/certgen.go > $tmpdir/env
. $tmpdir/env
oc process -p TLS_CRT="$TLS_CRT" -p TLS_KEY="$TLS_KEY" -p DOMAIN="$domain" -f routes.yaml | oc apply -f -
