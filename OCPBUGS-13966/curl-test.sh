#!/usr/bin/env bash

if [[ "$(oc project -q)" != "ocpbugs13966" ]]; then
    echo "Expecting current namespace to be \"ocpbugs13966\"."
    echo "Run: \"oc new-project ocpbugs13966\" first".
    exit 1
fi

routes=$(oc get routes -o jsonpath='{.items[*].metadata.name}')

if [[ -z "$routes" ]]; then
    echo "Error: no routes found in namespace $(oc project -q)."
    exit 1
fi

route_hostnames=$(oc get routes -o jsonpath='{.items[*].spec.host}')

for i in $(seq 1 ${1:-10}); do
    for host in $route_hostnames; do
	url="https://$host/test"
	result=$(curl -o /dev/null -k -L -s -w "%{http_code}\n" $url)
	echo "status $result for $url"
    done
done
