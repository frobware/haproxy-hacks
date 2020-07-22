#!/bin/bash

for i in $(seq 1 ${1:-10}); do
cat <<-EOF
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
