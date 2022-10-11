#!/usr/bin/env bash

set -eu

: "${DOMAIN:=int.frobware.com}"

. common.sh

for name in $(docker_pods | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    name=${name//_/-}
    echo "$name.$DOMAIN $port"
done | go run gen-mb-config.go "$@"
