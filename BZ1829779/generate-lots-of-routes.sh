#!/usr/bin/env bash

oc new-project lots-of-routes
./generate-pods.sh 1 | oc apply -f -
./generate-services.sh ${1:-100} | oc apply -f -
./expose-services.sh ${1:-100}
