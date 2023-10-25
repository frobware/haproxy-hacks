#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

domain=$(oc get ingresscontroller default -n openshift-ingress-operator -o jsonpath='{.status.domain}')

shard=$2

for i in $(seq 0 "${1:-10}"); do
    route_host="x-r${i}.${shard}.${domain}"
    cat <<-EOF
---
apiVersion: v1
kind: List
items:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld
      shard: ${shard}
    name: x-r${i}
  spec:
    host: ${route_host}
    port:
      targetPort: 8080
    to:
      kind: Service
      name: helloworld
    wildcardPolicy: None
EOF
done
