#!/usr/bin/env bash

set -eu

for i in $(seq ${1:-1} ${2:-10}); do
    oc process -o yaml -p ID="$i" -f pods.yaml | oc delete --ignore-not-found -f -
    [[ -n "${C:-}" ]] && oc process -o yaml -p ID="$i" -f pods.yaml | oc apply -f -
done

exit 0
