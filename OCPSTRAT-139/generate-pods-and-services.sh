#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first."
    exit 1
fi

cat <<-EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helloworld
  labels:
    app: helloworld
spec:
  replicas: 1
  selector:
    matchLabels:
      app: helloworld
  template:
    metadata:
      labels:
        app: helloworld
    spec:
      containers:
      - name: helloworld
        image: quay.io/openshift/origin-hello-openshift
        imagePullPolicy: IfNotPresent
        env:
        - name: RESPONSE
          value: "helloworld"
        - name: PORT
          value: "8080"
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          capabilities:
            drop:
            - ALL
          seccompProfile:
            type: RuntimeDefault
---
apiVersion: v1
kind: Service
metadata:
  name: helloworld
spec:
  selector:
    app: helloworld
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
EOF
