#!/usr/bin/env bash

set -eu

. common.sh

for name in $(docker_pods | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    name=${name//_/-}
    curl -4 -o /dev/null -k -L -s -w "${name} status %{http_code}\n" https://"$name.$(domain):$port/1024.html"
done
