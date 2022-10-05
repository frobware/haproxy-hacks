#!/usr/bin/env bash

for i in $(seq ${1:-1} ${2:-10}); do
cat <<-EOF
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: http-perf
    app: nginx-${i}
  name: http-perf-${i}
spec:
  containers:
  - image: quay.io/openshift-scale/nginx
    name: nginx
    resources:
      requests:
        memory: "10Mi"
        cpu: "10m"
    securityContext:
      allowPrivilegeEscalation: false
      runAsNonRoot: true
      capabilities:
        drop:
        - ALL
      seccompProfile:
        type: RuntimeDefault
  dnsPolicy: ClusterFirst
  restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: http-perf
  name: http-perf-${i}
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: 8443
  selector:
    app: nginx-${i}
  type: ${SERVICE_TYPE:-NodePort}
EOF
done
