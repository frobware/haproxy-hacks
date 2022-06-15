#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
oc adm policy remove-role-from-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
oc delete -n openshift-ingress-operator -f $thisdir/role.yaml
oc delete serviceaccount thanos
