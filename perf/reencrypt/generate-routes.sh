#!/usr/bin/env bash

set -eu

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

if [[ $? -ne 0 ]] || [[ -z $domain ]]; then
    echo "get domain failed" >&2
    exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "$tmpdir"' EXIT

go build -o $tmpdir/certgen ../certgen/certgen.go

for i in $(seq ${1:-1} ${2:-10}); do
    $tmpdir/certgen > $tmpdir/env
    . $tmpdir/env
    oc process -o yaml -p ID="$i" -p TLS_CRT="$TLS_CRT" -p TLS_CACRT="$TLS_CACRT" -p TLS_KEY="$TLS_KEY" -p DOMAIN="$domain" -f routes.yaml | oc delete --ignore-not-found -f -
    [[ -n "${C:-}" ]] && oc process -o yaml -p ID="$i" -p TLS_CRT="$TLS_CRT" -p TLS_CACRT="$TLS_CACRT" -p TLS_KEY="$TLS_KEY" -p DOMAIN="$domain" -f routes.yaml | oc apply -f -
done

exit 0

