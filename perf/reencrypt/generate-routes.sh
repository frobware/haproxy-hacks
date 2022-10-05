#!/usr/bin/env bash

: "${SHARD:="perf"}"
: "${NAMESPACE:="scale"}"
: "${DOMAIN:="ocp411.int.frobware.com"}"

for i in $(seq ${1:-1} ${2:-10}); do
    cat <<-EOF
kind: Route
apiVersion: route.openshift.io/v1
metadata:
  labels:
    type: perf
  name: http-perf-reencrypt-${i}
spec:
  host: http-perf-reencrypt-${i}-${NAMESPACE}.${SHARD}.${DOMAIN}
  port:
    targetPort: https
  tls:
    termination: reencrypt
    certificate: |
    key: |
  to:
    kind: Service
    name: http-perf-${i}
EOF
done
