#!/usr/bin/env bash

case "$(hostname)" in
    spicy*)
	container_name_prefix=docker_nginx_;
	;;
    *)
	container_name_prefix=docker-nginx-;
	;;
esac

host_ip=$(dig +search +short "$(hostname)")
domain=$(hostname -d)

if [[ -z "${domain}" ]]; then
    echo "error: no domain from hostname -d"
    exit 1
fi

declare -A BACKEND_CONTAINER_IDS=()
declare -A BACKEND_HTTPS_PORTS=()
declare -A BACKEND_HTTP_PORTS=()

for name in $(docker ps --no-trunc --filter name=^/${container_name_prefix} --format '{{.Names}}'); do
    https_port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    http_port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8080/tcp") 0).HostPort}}' "$name")"
    container_id="$(docker inspect --format='{{.Id}}' "$name")"
    name=${name//_/-}
    BACKEND_CONTAINER_IDS[$name]=$container_id
    BACKEND_HTTPS_PORTS[$name]=$https_port
    BACKEND_HTTP_PORTS[$name]=$http_port
done

function backend_names_sorted() {
    printf "%s\n" "${!BACKEND_HTTPS_PORTS[@]}" | sort -V
}

function dumparray() {
    declare -n __p="$1";

    for k in "${!__p[@]}"; do
	printf "%s=%s\n" "$k" "${__p[$k]}"
    done
}

# dumparray BACKEND_HTTP_PORTS
# dumparray BACKEND_HTTPS_PORTS
