#!/usr/bin/env bash

#!/usr/bin/env bash

for i in {1..1}; do
cat <<-EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helloworld-deployment-${i}
spec:
  replicas: 5
  selector:
    matchLabels:
      app: helloworld-deployment-${i}
  template:
    metadata:
      labels:
        app: helloworld-deployment-${i}
    spec:
      containers:
      - name: helloworld-container
        image: quay.io/openshift/origin-hello-openshift
        imagePullPolicy: Always
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
EOF
done

for i in {1..1}; do
cat <<-EOF
---
apiVersion: v1
kind: Service
metadata:
  name: helloworld-service-${i}
spec:
  selector:
    app: helloworld-deployment-${i}
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
EOF
done
