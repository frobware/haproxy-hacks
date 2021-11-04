#!/usr/bin/env bash

set -eu

: ${NPROXY:="1000"}
: ${WEIGHTS:="256"}
: ${ALGORITHMS:="leastconn random"} # compare these two by default

# https://github.com/sharkdp/hyperfine

hyperfine \
    -L nproxy "${NPROXY// /,}" \
    -L weight "${WEIGHTS// /,}" \
    -L algorithm "${ALGORITHMS// /,}" \
    --prepare './generate-haproxy-config.pl --balance-algorithm={algorithm} --nproxy={nproxy} --weight={weight} --output-dir=benchmark-config-algorithm-{algorithm}-nproxy-{nproxy}-weight-{weight}' \
    "$@" \
    'haproxy -c -f haproxy.config -C benchmark-config-algorithm-{algorithm}-nproxy-{nproxy}-weight-{weight}' \
