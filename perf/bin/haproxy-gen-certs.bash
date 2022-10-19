#!/usr/bin/env bash

set -eu

. common.sh

[[ -d "${1:?}/conf" ]] || {
    echo "$1/conf directory not found";
    exit 1
}

dest_cacrt="$(cat certs/reencrypt/destca.crt)"
tls_crt="$(cat certs/reencrypt/tls.crt)"
tls_key="$(cat certs/reencrypt/tls.key)"

for name in $(docker_pods | sort -V); do
    name=${name//_/-}
    printf "%s\n" "${dest_cacrt}" > "$1/router/cacerts/be_secure:${name}.pem"
    printf "%s\n%s\n" "$tls_key" "$tls_crt" > "$1/router/certs/be_secure:${name}.pem"
    echo "$1/router/certs/be_secure:${name}.pem ${name}.$(domain)" >> "$1/conf/cert_config.map"
done
