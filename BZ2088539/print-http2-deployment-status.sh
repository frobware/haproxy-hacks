#! /usr/bin/env bash

set -eu
set -o pipefail

for i in 46 47 48 49 410; do
    echo "Testing OCP $i"
    p="$HOME/src/github.com/frobware/infra/ocp${i}.int.frobware.com"
    PATH=$p:$PATH
    export KUBECONFIG="$p/ocp/auth/kubeconfig"
    oc version
    oc get deployment -n openshift-ingress -o yaml | grep -A2 HTTP2
    oc get pods -n openshift-ingress
done
