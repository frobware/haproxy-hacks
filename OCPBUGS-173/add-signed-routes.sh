#! /usr/bin/env bash

set -eu
set -o pipefail

: "${NAMESPACE:=default}"

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

if [[ $? -ne 0 ]] || [[ -z $domain ]]; then
    echo "get domain failed" >&2
    exit 1
fi

EDGE_TLS_CRT="$(cat ~/.acme.sh/websocket-edge-default.apps.ocp411.int.frobware.com/fullchain.cer)"
EDGE_TLS_KEY="$(cat ~/.acme.sh/websocket-edge-default.apps.ocp411.int.frobware.com/websocket-edge-default.apps.ocp411.int.frobware.com.key)"

REENCRYPT_TLS_CRT="$(cat ~/.acme.sh/websocket-reencrypt-default.apps.ocp411.int.frobware.com/fullchain.cer)"
REENCRYPT_TLS_KEY="$(cat ~/.acme.sh/websocket-reencrypt-default.apps.ocp411.int.frobware.com/websocket-reencrypt-default.apps.ocp411.int.frobware.com.key)"

oc process -o yaml \
   -p DOMAIN="$domain" \
   -p EDGE_TLS_CRT="$EDGE_TLS_CRT" \
   -p EDGE_TLS_KEY="$EDGE_TLS_KEY" \
   -p NAMESPACE="$NAMESPACE" \
   -p REENCRYPT_TLS_CRT="$REENCRYPT_TLS_CRT" \
   -p REENCRYPT_TLS_KEY="$REENCRYPT_TLS_KEY" \
   -f routes.yaml
