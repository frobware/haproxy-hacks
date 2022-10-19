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

declare -A -x BACKEND_CONTAINER_IDS
declare -A -x BACKEND_PORTS

for name in $(docker ps --no-trunc --filter name=^/${container_name_prefix} --format '{{.Names}}'); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    container_id="$(docker inspect --format='{{.Id}}' "$name")"
    name=${name//_/-}
    BACKEND_CONTAINER_IDS[$name]=$container_id
    BACKEND_PORTS[$name]=$port
done

function backend_names_sorted() {
    printf "%s\n" "${!BACKEND_PORTS[@]}" | sort -V
}
