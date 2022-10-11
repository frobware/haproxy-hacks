#!/usr/bin/env bash

set -eu

: "${DOMAIN:=int.frobware.com}"

for name in $(docker ps --no-trunc --filter name=^/docker-nginx- --format '{{.Names}}' | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    name=$(echo $name | sed 's/_/-/g')
    echo "$name.$DOMAIN $port"
done | go run gen-mb-config.go "$@"
