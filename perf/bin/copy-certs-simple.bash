#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

# This is the output path for the generated haproxy config.
: "${HAPROXY_CONFIG_DIR:=$(realpath "$PWD"/../haproxy-config)}"
HAPROXY_CONFIG_DIR=/tmp/${USER:-nouser}-haproxy-gen
export HAPROXY_CONFIG_DIR
echo "HAProxy configuration directory: $HAPROXY_CONFIG_DIR"

if [[ ! -f /tmp/haproxy-default.pem ]]; then
    "${thisdir}"/gen-domain-cert.bash
else
    echo -n "reusing existing cert: "; ls -l /tmp/haproxy-default.pem
fi

mkdir -p "${HAPROXY_CONFIG_DIR}"/router/{certs,cacerts,whitelists}
mkdir -p "${HAPROXY_CONFIG_DIR}"/{conf/.tmp,run,bin,log}
touch "${HAPROXY_CONFIG_DIR}"/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}

dest_cacrt="$(cat "${thisdir}"/certs/reencrypt/destca.crt)"
tls_crt="$(cat "${thisdir}"/certs/reencrypt/tls.crt)"
tls_key="$(cat "${thisdir}"/certs/reencrypt/tls.key)"

domain=localdomain

curl -s http://127.0.0.1:2000/backends | while read name hostaddr port; do
    if [[ $name =~ reencrypt ]]; then
	printf "%s\n" "${dest_cacrt}" > "${HAPROXY_CONFIG_DIR}/router/cacerts/be_secure:${name}.pem"
	printf "%s\n%s\n" "$tls_key" "$tls_crt" > "${HAPROXY_CONFIG_DIR}/router/certs/be_secure:${name}.pem"
	echo "${HAPROXY_CONFIG_DIR}/router/certs/be_secure:${name}.pem ${name}" >> "${HAPROXY_CONFIG_DIR}/conf/cert_config.map"
    fi
    if [[ $name =~ edge ]]; then
	printf "%s\n" "${dest_cacrt}" > "${HAPROXY_CONFIG_DIR}/router/cacerts/be_edge_http:${name}.pem"
	printf "%s\n%s\n" "$tls_key" "$tls_crt" > "${HAPROXY_CONFIG_DIR}/router/certs/be_edge_http:${name}.pem"
	echo "${HAPROXY_CONFIG_DIR}/router/certs/be_edge_http:${name}.pem ${name}" >> "${HAPROXY_CONFIG_DIR}/conf/cert_config.map"
    fi
done
