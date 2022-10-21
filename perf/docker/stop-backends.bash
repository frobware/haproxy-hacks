#!/usr/bin/env bash

for i in edge passthrough reencrypt http; do
    docker-compose -f "$i.yaml" down --remove-orphans --timeout 1
done
