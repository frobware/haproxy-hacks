#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

# oc create serviceaccount thanos
# oc describe serviceaccount thanos

oc apply -f role.yaml
oc apply -f rolebinding.yaml

SECRET=$(oc get secret -n openshift-ingress-operator | grep ingress-operator-token | head -n 1 | awk '{print $1 }')
oc process TOKEN=$SECRET -f triggerauthentication.yaml | oc apply -f -

# oc adm policy add-role-to-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
oc adm policy add-role-to-user thanos-metrics-reader -z ingress-operator --role-namespace=openshift-ingress-operator
