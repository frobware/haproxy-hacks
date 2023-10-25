#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

for i in $(seq "${1:-1}" "${2:-${MAX_SHARDS:-10}}"); do
    shard_name="shard${i}"
    while :; do
        rollout_status=$(oc get ingresscontroller/${shard_name} -n openshift-ingress-operator -o jsonpath='{.status.conditions[?(@.type=="Progressing")].status}')
        if [ "$rollout_status" == "False" ]; then
            echo "$shard_name: rollout complete."
            break
        else
            echo "$shard_name: rollout incomplete..."
            sleep 1
        fi
    done
done
