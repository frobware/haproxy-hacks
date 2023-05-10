#!/usr/bin/bash

set -eu

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

TLS_KEY="$(cat /home/aim/.acme.sh/router.int.frobware.com/router.int.frobware.com.key)"
TLS_CRT="$(cat /home/aim/.acme.sh/router.int.frobware.com/router.int.frobware.com.cer)"
TLS_CACRT="$(cat /home/aim/.acme.sh/router.int.frobware.com/ca.cer)"
DEST_CACRT="$(oc exec -n openshift-ingress -c router "$pod" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt)"

oc process DEST_CACRT="$DEST_CACRT" TLS_KEY="$TLS_KEY" TLS_CRT="$TLS_CRT" TLS_CACRT="$TLS_CACRT" -f ./destca-routes.yaml -o yaml | oc delete --ignore-not-found -f -
oc process DEST_CACRT="$DEST_CACRT" TLS_KEY="$TLS_KEY" TLS_CRT="$TLS_CRT" TLS_CACRT="$TLS_CACRT" -f ./destca-routes.yaml -o yaml | oc apply -f -
oc get routes
