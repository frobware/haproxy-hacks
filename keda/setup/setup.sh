#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
pushd $thisdir > /dev/null 2>&1

oc apply -f configmap.yaml
oc get -n openshift-ingress-operator serviceaccount/thanos && oc delete -n openshift-ingress-operator serviceaccount thanos
oc create -n openshift-ingress-operator serviceaccount thanos
oc describe -n openshift-ingress-operator serviceaccount thanos
oc apply -n openshift-ingress-operator -f role.yaml
secret=$(oc get secret -n openshift-ingress-operator | grep thanos-token | head -n 1 | awk '{print $1 }')
oc process TOKEN=$secret -f triggerauthentication.yaml | oc apply -n openshift-ingress-operator -f -
oc adm policy add-role-to-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
oc get triggerauthentications.keda.sh -o yaml | grep $secret

# add cluster-monitoring-view for cross-namespace queries; scaled object must target port 9091
oc adm policy add-cluster-role-to-user cluster-monitoring-view -z thanos

#$ oc get secrets -n openshift-monitoring thanos-querier-token-gqstv -o yaml |grep token
# secret=$(oc get secrets -n openshift-monitoring thanos-querier-token-gqstv -o yaml | token: | head -n 1 | awk '{print $1 }')
