#!/usr/bin/env bash

cat <<-EOF
---
apiVersion: v1
kind: Pod
metadata:
  name: helloworld-1
  labels:
    app: helloworld-1
spec:
  containers:
  - name: helloworld-1
    image: quay.io/openshift/origin-hello-openshift
    imagePullPolicy: Always
    env:
    - name: RESPONSE
      value: "helloworld-1"
    - name: PORT
      value: "8080"
---
apiVersion: v1
kind: Service
metadata:
  name: helloworld-1
spec:
  selector:
    app: helloworld-1
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
EOF
