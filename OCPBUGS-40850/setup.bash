#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs40850" ]]; then
    echo "Expecting current namespace to be \"ocpbugs40850\"."
    echo "Run: \"oc new-project ocpbugs40850\" first".
    exit 1
fi

oc delete routes --all
oc delete --ignore-not-found -f ./manifests
oc apply -f ./manifests
./gen-routes.bash "${1:-3}" | oc apply -f -
