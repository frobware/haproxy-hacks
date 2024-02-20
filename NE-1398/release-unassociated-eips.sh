#!/usr/bin/env bash

set -euo pipefail

# List all unassociated EIPs and their IP addresses.
unassociated_eips=$(aws ec2 describe-addresses \
                        --query 'Addresses[?AssociationId==`null`].[PublicIp, AllocationId]' \
                        --output text)

if [ -z "$unassociated_eips" ]; then
    echo "No unassociated EIPs found."
    exit 0
fi

# Release each unassociated EIP and print IP address and Allocation ID.
while read -r public_ip allocation_id; do
    if ! aws ec2 release-address --allocation-id "$allocation_id"; then
        echo "Failed to release EIP $public_ip with Allocation ID: $allocation_id" >&2
    else
        echo "Released EIP $public_ip with Allocation ID: $allocation_id"
    fi
done <<< "$unassociated_eips"

echo "All unassociated EIPs have been released."
