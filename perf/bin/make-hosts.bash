#!/usr/bin/env bash

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"

: "${HOSTIP:=$(dig +short $(hostname))}"
: "${DOMAIN:=$domain}"

# For some sanity invoke this as:
#
# $ ./bin/make-hosts | sort -k2 -V > hosts

for name in $(backend_names_sorted); do
    for termination_type in edge http passthrough reencrypt; do
	echo "$HOSTIP ${name}-${termination_type}.${domain}"
    done
done
