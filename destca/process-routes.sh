#!/usr/bin/bash

set -eu
trap "rm -f /tmp/certgen.$$" EXIT
GO111MODULE=off go run certgen.go > /tmp/certgen.$$
source /tmp/certgen.$$
oc process TLS_KEY="$TLS_KEY" TLS_CRT="$TLS_CRT" TLS_CACRT="$TLS_CACRT" -f ./destca-routes.yaml | oc apply -f -
oc get routes 
