#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpstrat139" ]]; then
    echo "Expecting current namespace to be \"ocpstrat139\"."
    echo "Run: \"oc new-project ocpstrat139\" first".
    exit 1
fi

domain=$(oc get ingresscontroller default -n openshift-ingress-operator -o jsonpath='{.status.domain}')

shards=$(oc get ingresscontrollers -n openshift-ingress-operator -o jsonpath='{.items[*].metadata.name}')
read -ra shard_array <<< "$shards"

# Filter out the 'default' ingresscontroller.
shard_array=("${shard_array[@]/default/}")

# Remove empty elements.
shard_array=($(printf "%s\n" "${shard_array[@]}" | grep -v '^$'))

num_shards=${#shard_array[@]}

for i in $(seq 0 "${1:-10}"); do
    shard_index=$(( i % num_shards ))
    current_shard=${shard_array[$shard_index]}
    route_host="r${i}.${current_shard}.${domain}"
    cat <<-EOF
---
apiVersion: v1
kind: List
items:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld
      shard: ${current_shard}
    name: r${i}
  spec:
    host: ${route_host}
    port:
      targetPort: 8080
    to:
      kind: Service
      name: helloworld
    wildcardPolicy: None
EOF
done
