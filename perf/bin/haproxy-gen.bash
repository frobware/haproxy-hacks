#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
pushd "$(realpath $thisdir)" >/dev/null

. common.sh

# This is the output path for the generaed haproxy config.
: "${HAPROXY_CONFIG_DIR:=$(realpath $PWD/../haproxy-config)}"
export HAPROXY_CONFIG_DIR
echo "HAProxy configuration directory: $HAPROXY_CONFIG_DIR"

if [[ -z "$(domain)" ]]; then
    echo "error: no domain from hostname -d"
    exit 1
fi

if [[ ! -f /tmp/haproxy-default.pem ]]; then
    ./gen-haproxy-cert.bash
else
    echo -n "reusing existing cert: "; ls -l /tmp/haproxy-default.pem
fi

rm -rf ${HAPROXY_CONFIG_DIR}
mkdir -p ${HAPROXY_CONFIG_DIR}/router/{certs,cacerts,whitelists}
mkdir -p ${HAPROXY_CONFIG_DIR}/{conf/.tmp,run,bin,log}
touch ${HAPROXY_CONFIG_DIR}/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}
cp error-page-404.http ${HAPROXY_CONFIG_DIR}/conf
cp error-page-503.http ${HAPROXY_CONFIG_DIR}/conf
./haproxy-gen-preamble.bash > ${HAPROXY_CONFIG_DIR}/conf/haproxy.config
./haproxy-gen-backends.bash >> ${HAPROXY_CONFIG_DIR}/conf/haproxy.config
./haproxy-gen-os_edge_reencrypt_be.map > ${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map
./haproxy-gen-certs.bash ${HAPROXY_CONFIG_DIR}
