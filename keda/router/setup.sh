#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

oc create -n openshift-ingress serviceaccount thanos

oc apply -n openshift-ingress -f $thisdir/role.yaml
oc apply -n openshift-ingress -f $thisdir/rolebinding.yaml

SECRET=$(oc get secret -n openshift-ingress | grep thanos-token | head -n 1 | awk '{print $1 }')
oc process TOKEN=$SECRET -f $thisdir/triggerauthentication.yaml | oc apply -n openshift-ingress -f -

oc adm policy add-role-to-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress
