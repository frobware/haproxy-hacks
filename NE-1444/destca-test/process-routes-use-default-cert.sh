#!/usr/bin/env bash

set -eu

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

if [[ $# -ne 1 ]]; then
    echo "Specify a route.yaml"
    exit 1
fi

route_spec=$1

oc process -f $route_spec -o yaml | oc delete --ignore-not-found -f -
oc process -f $route_spec -o yaml | oc apply -f -
