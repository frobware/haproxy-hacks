#!/usr/bin/env bash

oc create -f deployment.yaml
oc logs deployment.apps/test-app
oc create serviceaccount thanos
oc describe serviceaccount thanos
oc create -f triggerauthentication.yaml
oc create -f role.yaml
oc adm policy add-role-to-user thanos-metrics-reader -z thanos --role-namespace=openshift-ingress-operator
