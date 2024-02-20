#!/usr/bin/env bash

set -euo pipefail

# List all EIPs, highlighting unassociated EIPs.
aws ec2 describe-addresses --output json | \
    jq -r '.Addresses[] | select(.AssociationId == null) | [.PublicIp + " " + .AllocationId] + (reduce .Tags[]? as $tag ([]; . + ["\($tag.Key)=\($tag.Value)"])) | join(" ")'
