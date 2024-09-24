#!/usr/bin/env bash

set -eu

# Number of routes to create for each type (default is 3).
N=${1:-3}

# Arrays for termination types and corresponding targetPorts.
termination_types=("edge" "reencrypt" "passthrough")
target_ports_single_te=("single-te" "single-te-tls" "single-te-tls")
target_ports_duplicate_te=("dup-te" "dup-te-tls" "dup-te-tls")

for i in $(seq 1 "$N"); do
    for j in "${!termination_types[@]}"; do
        termination=${termination_types[$j]}
        targetPort=${target_ports_single_te[$j]}
        cat <<-EOF
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: ocpbugs40850-single-te-${termination}-${i}
  labels:
    app: ocpbugs40850-test
spec:
  port:
    targetPort: ${targetPort}
  tls:
    termination: ${termination}
    insecureEdgeTerminationPolicy: Redirect
  to:
    kind: Service
    name: ocpbugs40850-test
    weight: 100
  wildcardPolicy: None
EOF
    done
    for j in "${!termination_types[@]}"; do
        termination=${termination_types[$j]}
        targetPort=${target_ports_duplicate_te[$j]}
        cat <<-EOF
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: ocpbugs40850-dup-te-${termination}-${i}
  labels:
    app: ocpbugs40850-test
spec:
  port:
    targetPort: ${targetPort}
  tls:
    termination: ${termination}
    insecureEdgeTerminationPolicy: Redirect
  to:
    kind: Service
    name: ocpbugs40850-test
    weight: 100
  wildcardPolicy: None
EOF
    done
done
