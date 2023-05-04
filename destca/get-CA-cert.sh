#!/usr/bin/env bash

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [ -z "$pod" ]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

oc exec -n openshift-ingress -c router $pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt
