#!/usr/bin/env bash

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first."
    exit 1
fi

# Remove label from all existing routes.
echo "Removing type=default labels from all existing routes..."
oc get routes --all-namespaces -o custom-columns="NAMESPACE:.metadata.namespace,NAME:.metadata.name" --no-headers | while read -r namespace route_name; do
    oc label route "$route_name" -n "$namespace" type-
done

# Revert changes to the default ingress controller.
echo "Reverting changes to the default ingress controller..."
oc patch ingresscontroller/default -n openshift-ingress-operator --type='json' -p='[{"op": "remove", "path": "/spec/routeSelector"}]'

while :; do
    rollout_status=$(oc get ingresscontroller/default -n openshift-ingress-operator -o jsonpath='{.status.conditions[?(@.type=="Progressing")].status}')

    if [ "$rollout_status" == "False" ]; then
        echo "ingresscontroller rollout reverted successfully."
        break
    else
        echo "Waiting for ingresscontroller rollout to revert..."
        sleep 1
    fi
done
