#! /usr/bin/env bash

set -eu
set -o pipefail

for i in 46 47 48 49 410; do
    echo "Testing OCP $i"
    p="$HOME/src/github.com/frobware/infra/ocp${i}.int.frobware.com"
    PATH=$p:$PATH
    export KUBECONFIG="$p/ocp/auth/kubeconfig"
    echo $KUBECONFIG
    oc version
    oc -n openshift-ingress-operator patch ingresscontroller/default --type=merge --patch='{"spec":{"logging":{"access":{"destination":{"type":"Container"}}}}}'
    oc -n openshift-ingress-operator scale --replicas=1 ingresscontroller/default
    oc apply -f deployment.yaml
    ./create-routes.sh
    oc get routes
done
