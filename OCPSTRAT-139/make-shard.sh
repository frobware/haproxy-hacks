#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

domain=$(oc get ingresscontroller default -n openshift-ingress-operator -o jsonpath='{.status.domain}')

for i in $(seq "${1:-1}" "${2:-${MAX_SHARDS:-10}}"); do
cat <<-EOF
---
apiVersion: operator.openshift.io/v1
kind: IngressController
metadata:
  name: shard${i}
  namespace: openshift-ingress-operator
spec:
  domain: shard${i}.$domain
  namespace:
    name: openshift-ingress-operator
  routeSelector:
    matchLabels:
      shard: shard${i}
  replicas: 1
  logging:
    access:
      destination:
        type: Container
      logEmptyRequests: Log
EOF
done

