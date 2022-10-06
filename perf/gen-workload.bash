#!/usr/bin/env bash

set -eu

: "${DOMAIN:=int.frobware.com}"

for name in $(docker ps --no-trunc --filter name=^/docker_nginx_ --format '{{.Names}}' | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    container_id="$(docker inspect --format='{{.Id}}' "$name")"
    echo "$name.$DOMAIN"
done | go run gen-mb-config.go "$@"
