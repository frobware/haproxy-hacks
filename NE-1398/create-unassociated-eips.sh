#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <number-of-EIPs-to-allocate>" >&2
    exit 1
fi

num_eips=$1

: "${NE1398_TAG_KEY:=NE1398-${USER}}"
: "${NE1389_TAG_VALUE:=$(./cluster-name.sh)}"

for ((i=0; i<num_eips; i++)); do
    if ! eip_info=$(aws ec2 allocate-address \
                        --domain vpc --query 'AllocationId' \
                        --output text); then
        echo "Failed to allocate EIP." >&2
        exit 1
    fi

    allocation_id=$eip_info
    echo "EIP allocated successfully. Allocation ID: $allocation_id"

    name="${NE1389_TAG_VALUE}-EIP-${RANDOM}"
    if ! aws ec2 create-tags \
         --resources "$allocation_id" \
         --tags "Key=Name,Value=${name}" "Key=$NE1398_TAG_KEY,Value=$NE1389_TAG_VALUE"; then
        echo "Failed to tag EIP Allocation ID: $allocation_id." >&2
        exit 1
    fi

    echo "EIP Allocation ID: $allocation_id tagged successfully with Name=$name $NE1398_TAG_KEY=$NE1389_TAG_VALUE"
done
