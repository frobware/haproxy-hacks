#!/usr/bin/env bash

for i in $(seq ${1:-1} ${2:-1}); do
cat <<-EOF
---
apiVersion: v1
kind: List
items:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-1
      type: blueprint_route_labels
    name: helloworld-${i}-edge
  spec:
    port:
      targetPort: 8080
    tls:
      termination: edge
      insecureEdgeTerminationPolicy: Redirect
    to:
      kind: Service
      name: helloworld-service-1
      weight: 100
    wildcardPolicy: None
EOF
done
