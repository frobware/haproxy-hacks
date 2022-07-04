#! /usr/bin/env bash

set -eu
set -o pipefail
declare -A status

for i in 46 47 48 49 410; do
    echo "Testing OCP $i"
    p="$HOME/src/github.com/frobware/infra/ocp${i}.int.frobware.com"
    PATH=$p:$PATH
    export KUBECONFIG="$p/ocp/auth/kubeconfig"
    set +e
    curl_output="$(curl -s -k -L "https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz")"
    curl_status=$?
    if [[ $curl_status -ne 0 ]]; then
        status["OCP-$i"]="failed: status=$curl_status; $curl_output"
    fi
    set -e
done

for i in "${!status[@]}"; do
    echo "${i}=${status[$i]}"
done
