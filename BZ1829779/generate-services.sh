#!/bin/bash

for i in $(seq 1 ${1:-10}); do
cat <<-EOF
---
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: helloworld-${i}
  spec:
    selector:
      app: helloworld-${i}
    ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-${i}
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
      name: helloworld-${i}
      weight: 100
    wildcardPolicy: None
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-${i}
    name: helloworld-${i}-insecure
  spec:
    port:
      targetPort: 8080
    to:
      kind: Service
      name: helloworld-${i}
      weight: 100
    wildcardPolicy: None
EOF
done
