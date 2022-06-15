#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
SECRET=$(oc get secret -n openshift-ingress-operator | grep ingress-operator-token | head -n 1 | awk '{print $1 }')
oc process TOKEN=$SECRET -f $thisdir/triggerauthentication.yaml | oc replace --force -f -



