#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs14914" ]]; then
    echo "Expecting current namespace to be \"ocpbugs14914\"."
    echo "Run: \"oc new-project ocpbugs14914\" first".
    exit 1
fi

oc scale --replicas=1 -n openshift-ingress-operator ingresscontroller/default
oc rollout status deployment/router-default -n openshift-ingress

npods=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default --no-headers | wc -l)

if [[ $npods -ne 1 ]]; then
    echo "expected 1 router pod, get $npods pods."
    exit 1
fi

oc delete --all routes
oc delete --all services
oc delete --all deployments

image="$(oc get deployment ingress-operator -n openshift-ingress-operator -o=jsonpath='{.spec.template.spec.containers[?(@.name=="ingress-operator")].image}')"
oc process -f manifests/deployment-template.yaml -p IMAGE=$image | oc apply -f -
oc apply -f manifests/passthrough-route.yaml
oc apply -f manifests/service.yaml

oc rollout status deployment/ocpbugs14914-test
oc get all
