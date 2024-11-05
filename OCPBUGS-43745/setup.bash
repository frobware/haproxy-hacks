#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs43745" ]]; then
    echo "Expecting current namespace to be \"ocpbugs43745\"."
    echo "Run: \"oc new-project ocpbugs43745\" first".
    exit 1
fi

oc delete --ignore-not-found routes --all
oc delete --ignore-not-found services --all
oc delete --ignore-not-found rc --all
oc delete --ignore-not-found -f ./manifests
oc apply -f ./manifests
sleep 3

oc get replicationcontrollers -l name=web-server-rc1 -o wide
oc get replicationcontrollers -l name=web-server-rc2 -o wide
oc get pods -l name=web-server-rc1 -o wide
oc get pods -l name=web-server-rc2 -o wide
