#!/usr/bin/env bash

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

oc delete deployments --all
oc delete services --all
oc delete routes --all

#./make-shards.sh | oc --ignore-not-found=true delete -f -
