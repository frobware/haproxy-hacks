#! /usr/bin/env bash

set -eu
set -o pipefail

for i in 46 47 48 49 410; do
    echo "Testing OCP $i"
    p="$HOME/src/github.com/frobware/infra/ocp${i}.int.frobware.com"
    PATH=$p:$PATH
    export KUBECONFIG="$p/ocp/auth/kubeconfig"
    oc version
    oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=${1:-true}
done

sleep 90 # how to wait for the ingresscontroller and deployment update?
