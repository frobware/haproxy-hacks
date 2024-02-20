#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "usage: <VPC-ID>" >&2
    exit 1
fi

vpc_id=$1; shift
subnets=$(aws ec2 describe-subnets \
              --filters "Name=vpc-id,Values=$vpc_id" \
              --query 'Subnets[*].SubnetId' \
              --output text)

if [ -z "$subnets" ]; then
    echo "No subnets found in VPC $vpc_id." >&2
    exit 1
fi

subnet_ids=($subnets)

# Define some known tags so we can find stuff in the console.
NE1398_TAG_KEY="NE1398-${USER}"
NE1389_TAG_VALUE="$(./cluster-name.sh)"

# Allocate an EIP for each subnet and print the allocation details.
for subnet_id in "${subnet_ids[@]}"
do
    if ! eip_allocation=$(aws ec2 allocate-address \
                              --domain vpc --query '[PublicIp, AllocationId]' \
                              --output text); then
        echo "Failed to allocate EIP for subnet $subnet_id." >&2
        exit 1
    fi
    read -r public_ip allocation_id <<< "$eip_allocation"
    echo "Subnet ID: $subnet_id, EIP: $public_ip, Allocation ID: $allocation_id"
done
