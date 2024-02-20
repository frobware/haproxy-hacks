#!/usr/bin/env bash

set -euo pipefail

cluster_name="$(./cluster-name.sh)"
tag_key="kubernetes.io/cluster/${cluster_name}"
tag_value="owned"

vpc_id=$(aws ec2 describe-instances \
             --filters "Name=tag:${tag_key},Values=${tag_value}" \
             --query 'Reservations[0].Instances[0].VpcId' \
             --output text)

if [ "$vpc_id" = "None" ]; then
    echo "No instances found with the specified tag, or unable to determine VPC ID." >&2
    exit 1
fi

echo "$vpc_id"
