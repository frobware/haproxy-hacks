#! /usr/bin/env bash

set -eu

: "${IMAGE:=registry.int.frobware.com/aim/ocpbugs-173-server}"
oc process -p IMAGE="${IMAGE}" -f ./server/deployment.yaml | oc delete --ignore-not-found -f -
oc process -p IMAGE="${IMAGE}" -f ./server/deployment.yaml | oc apply -f -
./add-signed-routes.sh | oc delete --ignore-not-found -f -
./add-signed-routes.sh | oc apply -f -
