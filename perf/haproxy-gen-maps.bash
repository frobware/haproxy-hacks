#!/usr/bin/env bash

for name in $(docker ps --no-trunc --filter name=^/docker_nginx_ --format '{{.Names}}' | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    container_id="$(docker inspect --format='{{.Id}}' "$name")"

    echo "
${name}.int.frobware.com\.?(:[0-9]+)?(/.*)?$ be_secure:${name}.int.frobware.com
"
done
