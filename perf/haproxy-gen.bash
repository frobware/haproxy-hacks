#!/usr/bin/env bash

set -eu

: "${HAPROXY_CONFIG_DIR:=$(realpath $PWD/haproxy)}"
export HAPROXY_CONFIG_DIR

rm -rf ${HAPROXY_CONFIG_DIR}
mkdir -p ${HAPROXY_CONFIG_DIR}/router/{certs,cacerts,whitelists}
mkdir -p ${HAPROXY_CONFIG_DIR}/{conf/.tmp,run,bin,log}
touch ${HAPROXY_CONFIG_DIR}/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}
cp certs/default.pem ${HAPROXY_CONFIG_DIR}/router/certs/default.pem
cp error-page-404.http ${HAPROXY_CONFIG_DIR}/conf
cp error-page-503.http ${HAPROXY_CONFIG_DIR}/conf
./haproxy-gen-preamble.bash > ${HAPROXY_CONFIG_DIR}/conf/haproxy.config
./haproxy-gen-backends.bash >> ${HAPROXY_CONFIG_DIR}/conf/haproxy.config
./haproxy-gen-os_edge_reencrypt_be.map > ${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map
./haproxy-gen-certs.bash ${HAPROXY_CONFIG_DIR}
