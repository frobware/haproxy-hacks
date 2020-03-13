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
---
apiVersion: v1
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
EOF
done
