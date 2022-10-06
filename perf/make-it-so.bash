#!/usr/bin/env bash

rm -rf /tmp/lib/haproxy
mkdir -p /tmp/lib/haproxy/router/{certs,cacerts,whitelists}
mkdir -p /tmp/lib/haproxy/{conf/.tmp,run,bin,log}
touch /tmp/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}
cp ~/domain.pem /tmp/lib/haproxy/router/certs/default.pem
./haproxy-gen.pl > /tmp/lib/haproxy/conf/haproxy.config
./haproxy-gen-backends.bash >> /tmp/lib/haproxy/conf/haproxy.config
# ./haproxy-gen-os_tcp_be.map > /tmp/lib/haproxy/conf/os_tcp_be.map
./haproxy-gen-os_edge_reencrypt_be.map > /tmp/lib/haproxy/conf/os_edge_reencrypt_be.map
./haproxy-gen-certs.bash /tmp/lib/haproxy
