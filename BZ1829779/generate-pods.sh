#!/bin/bash

for i in $(seq 1 ${1:-10}); do
cat <<-EOF
---
apiVersion: v1
kind: Pod
metadata:
  name: helloworld-${i}
  labels:
    app: helloworld-${i}
spec:
  containers:
  - name: helloworld-${i}
    image: openshift/hello-openshift
    imagePullPolicy: IfNotPresent
    env:
    - name: RESPONSE
      value: "helloworld-${i}"
    - name: PORT
      value: "8080"
EOF
done
