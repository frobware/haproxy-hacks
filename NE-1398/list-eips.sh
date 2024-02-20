#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "usage: $0 <VPC-ID>" >&2
    exit 1
fi

vpc_id=$1; shift

# List all ENIs in the specified VPC.
eni_ids=$(aws ec2 describe-network-interfaces \
              --filters "Name=vpc-id,Values=$vpc_id" \
              --query 'NetworkInterfaces[*].NetworkInterfaceId' \
              --output text)

if [ -z "$eni_ids" ]; then
    echo "No Elastic Network Interfaces (ENIs) found in VPC $vpc_id." >&2
    exit 1
fi

# For each ENI, check if there is an EIP associated, list it, and
# enumerate any tags.
for eni_id in $eni_ids; do
    aws ec2 describe-addresses \
        --filters "Name=network-interface-id,Values=$eni_id" \
        --output json | jq -r '.Addresses[] | [.PublicIp + " " + .AllocationId] + (reduce .Tags[] as $tag ([]; . + ["\($tag.Key)=\($tag.Value)"])) | join(" ")'
done
