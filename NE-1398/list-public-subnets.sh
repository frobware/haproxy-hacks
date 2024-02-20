#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <VPC-ID>" >&2
    exit 1
fi

vpc_id=$1

# Get the Internet Gateway associated with the VPC.
igws=$(aws ec2 describe-internet-gateways \
           --filters "Name=attachment.vpc-id,Values=$vpc_id" \
           --query 'InternetGateways[*].InternetGatewayId' \
           --output json | jq -r '.[]')

public_subnets=()

for igw in $igws; do
    # Look for route tables that have routes to the Internet Gateway.
    subnet_ids=$(aws ec2 describe-route-tables \
                     --query "RouteTables[?Routes[?GatewayId=='$igw']].Associations[].SubnetId" \
                     --output json | jq -r '.[] | select(. != null)')
    for subnet_id in $subnet_ids; do
        public_subnets+=("$subnet_id")
    done
done

echo "${public_subnets[@]}" | tr ' ' '\n' | sort -u
