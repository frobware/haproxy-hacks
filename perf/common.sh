#!/usr/bin/env bash

docker_pod_prefix=docker-nginx-

case "$(hostname)" in
    spicy*) docker_pod_prefix=docker_nginx_;;
esac

function docker_pods() {
    docker ps --no-trunc --filter name=^/${docker_pod_prefix} --format '{{.Names}}'
}

function domain() {
    local h=$(hostname)
    echo "${h#*.*}"
}
