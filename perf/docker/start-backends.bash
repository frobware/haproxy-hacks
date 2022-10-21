#!/usr/bin/env bash

#set -eu

for i in edge passthrough reencrypt http; do
    docker-compose -f "$i.yaml" up --detach --no-color --timeout 1 --scale "nginx_$i"=${1:-100}
done


