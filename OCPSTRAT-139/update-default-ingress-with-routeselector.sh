#!/usr/bin/env bash

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

# This script modifies the default OpenShift ingress controller to use
# a route selector. It also labels all existing routes so that they
# are selected by the new route selector.

# Label all existing routes with type=default, to be used by the
# default ingress controller's route selector.
echo "Labelling all existing routes to use the routeSelector on the default ingress controller..."
oc get routes --all-namespaces -o custom-columns="NAMESPACE:.metadata.namespace,NAME:.metadata.name" --no-headers | while read -r namespace route_name; do
  oc label route "$route_name" -n "$namespace" type=default --overwrite
done

# Patch the default ingress controller to use the new route selector.
echo "Patching the default ingress controller to use the new route selector..."
oc patch ingresscontroller/default -n openshift-ingress-operator --type='json' -p='[{"op": "add", "path": "/spec/routeSelector", "value": {"matchLabels": {"type": "default"}}}]'

while :; do
    rollout_status=$(oc get ingresscontroller/default -n openshift-ingress-operator -o jsonpath='{.status.conditions[?(@.type=="Progressing")].status}')

    if [ "$rollout_status" == "False" ]; then
	echo "ingresscontroller rollout complete."
	break
    else
	echo "Waiting for ingresscontroller rollout to complete..."
	sleep 1
    fi
done
