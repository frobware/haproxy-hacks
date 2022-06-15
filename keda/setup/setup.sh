#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

oc apply -f $thisdir/configmap.yaml
oc delete -n openshift-ingress-operator serviceaccount thanos
sleep 5
oc create -n openshift-ingress-operator serviceaccount thanos
oc describe -n openshift-ingress-operator serviceaccount thanos
oc apply -n openshift-ingress-operator -f $thisdir/role.yaml
#oc apply -n openshift-ingress-operator -f $thisdir/rolebinding.yaml
secret=$(oc get secret -n openshift-ingress-operator | grep thanos-token | head -n 1 | awk '{print $1 }')
oc process TOKEN=$secret -f $thisdir/triggerauthentication.yaml | oc apply -n openshift-ingress-operator -f -
oc adm policy add-role-to-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
oc adm policy add-role-to-user thanos-metrics-reader -z ingress-operator --role-namespace=openshift-ingress-operator
oc get triggerauthentications.keda.sh -o yaml | grep thanos-token

