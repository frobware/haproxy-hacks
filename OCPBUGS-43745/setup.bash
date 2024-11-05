#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs43745" ]]; then
    echo "Expecting current namespace to be \"ocpbugs43745\"."
    echo "Run: \"oc new-project ocpbugs43745\" first".
    exit 1
fi

oc delete routes --all
oc delete --ignore-not-found -f ./manifests
oc apply -f ./manifests
