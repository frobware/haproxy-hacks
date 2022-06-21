#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
pushd $thisdir > /dev/null 2>&1

oc adm policy remove-role-from-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
oc delete -n openshift-ingress-operator -f role.yaml
oc delete serviceaccount thanos
