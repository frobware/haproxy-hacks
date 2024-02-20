#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "usage: $0 <VPC-ID>" >&2
    exit 1
fi

vpc_id=$1; shift

if ! subnets=$(aws ec2 describe-subnets \
                   --filters "Name=vpc-id,Values=$vpc_id" \
                   --query 'Subnets[*].SubnetId' \
                   --output json | jq -r '.[]'); then
    echo "failed to describe-subnets" >&2
    exit 1
fi

if [ -z "$subnets" ]; then
    echo "No subnets found in VPC $vpc_id." >&2
    exit 1
fi

echo "$subnets"

subnet_ids=($subnets)
for subnet_id in "${subnet_ids[@]}"; do
    aws ec2 describe-subnets \
        --subnet-ids "$subnet_id" \
        --query 'Subnets[*].{ID:SubnetId,CIDR:CidrBlock,Zone:AvailabilityZone,Tags:Tags}' \
        --output table
done
