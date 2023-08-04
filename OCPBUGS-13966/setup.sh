#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs13966" ]]; then
    echo "Expecting current namespace to be \"ocpbugs13966\"."
    echo "Run: \"oc new-project ocpbugs13966\" first".
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

make -C server

pushd manifests
oc apply -f service.yaml
oc apply -f deployment.yaml
oc apply -f passthrough-route.yaml
popd

oc rollout status deployment/ocpbugs13966-test
oc get all
