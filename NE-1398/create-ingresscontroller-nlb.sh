#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "usage: $0 <ingress-controller-name>" >&2
    exit 1
fi

name="$1"; shift;
domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o json | jq -r '.status.domain')

if [[ $? -ne 0 ]] || [[ -z $domain ]]; then
    echo "failed to get cluster domain" >&2
    exit 1
fi

oc process -o yaml \
   -p DOMAIN="$domain" \
   -p NAME="$name" \
   -f ./ingresscontroller-nlb.yaml | oc apply -n openshift-ingress-operator -f -
