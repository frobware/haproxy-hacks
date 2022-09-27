#!/bin/bash

for i in $(seq 1 ${1:-10}); do
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
    name: helloworld-${i}-edge
  spec:
    port:
      targetPort: 8080
    tls:
      termination: edge
      insecureEdgeTerminationPolicy: Redirect
      key: |-
      certificate: |-
    to:
      kind: Service
      name: helloworld-1
      weight: 100
    wildcardPolicy: None
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-1
    name: helloworld-${i}-insecure
  spec:
    port:
      targetPort: 8080
    to:
      kind: Service
      name: helloworld-1
      weight: 100
    wildcardPolicy: None
EOF
done
