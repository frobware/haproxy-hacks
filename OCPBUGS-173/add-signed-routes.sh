#! /usr/bin/env bash

set -eu
set -o pipefail

: "${NAMESPACE:=default}"

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

if [[ $? -ne 0 ]] || [[ -z $domain ]]; then
    echo "get domain failed" >&2
    exit 1
fi

TLS_CRT="$(cat ~/.acme.sh/reproducer-default.apps.ocp411.int.frobware.com/reproducer-default.apps.ocp411.int.frobware.com.cer)"
TLS_KEY="$(cat ~/.acme.sh/reproducer-default.apps.ocp411.int.frobware.com/reproducer-default.apps.ocp411.int.frobware.com.key)"

oc process -o yaml -p NAMESPACE="$NAMESPACE" -p TLS_CRT="$TLS_CRT" -p TLS_KEY="$TLS_KEY" -p DOMAIN="$domain" -f routes.yaml
