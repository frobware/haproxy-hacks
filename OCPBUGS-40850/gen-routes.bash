#!/usr/bin/env bash

set -eu

# Number of routes to create for each type (default is 3).
N=${1:-10}

# Bash associative array to map termination types to target ports.
declare -A target_ports=(
    ["edge"]="http"       # Edge termination targets HTTP (8080)
    ["reencrypt"]="https" # Reencrypt termination targets HTTPS (8443)
    ["passthrough"]="https" # Passthrough termination targets HTTPS (8443)
)

# Arrays for termination types.
termination_types=("edge" "reencrypt" "passthrough")

for i in $(seq 1 "$N"); do
    for termination in "${termination_types[@]}"; do
        target_port="${target_ports[$termination]}"  # Get the appropriate target port from the map

        # Generate YAML for the route
        cat <<-EOF
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: ${termination}-${i}
  labels:
    app: ocpbugs40850-test
spec:
  port:
    targetPort: "${target_port}"  # Use the mapped target port
  tls:
    termination: ${termination}
    insecureEdgeTerminationPolicy: Redirect  # Redirect http to https
  to:
    kind: Service
    name: ocpbugs40850-test
    weight: 100
  wildcardPolicy: None
EOF
    done
done
