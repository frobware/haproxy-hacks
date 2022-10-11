#!/usr/bin/env bash

set -eu

: "${DOMAIN:=int.frobware.com}"

for name in $(docker ps --no-trunc --filter name=^/docker-nginx- --format '{{.Names}}' | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    name=$(echo $name | sed 's/_/-/g')
    echo "mkdir -p /tmp/$name; wrk2 -R 10000 -c 200 --threads 1 --timeout 30s https://$name.${DOMAIN}:8443/1024.html > /tmp/$name/result 2>&1 &"
done
