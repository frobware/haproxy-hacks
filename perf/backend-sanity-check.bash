#!/usr/bin/env bash

set -eu

. common.sh

for name in $(docker_pods | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    name=${name//_/-}
    curl -o /dev/null -s -k --connect-timeout 2 https://"$name.$(domain):$port/1024.html"
    echo "$name.$(domain) $port OK"
done
