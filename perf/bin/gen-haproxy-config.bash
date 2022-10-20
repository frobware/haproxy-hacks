#!/usr/bin/env bash

set -eu

thisdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
. "${thisdir}/common.sh"
: "${MAXCONN:=0}"
: "${NBTHREAD:=4}"

# This is the output path for the generated haproxy config.
: "${HAPROXY_CONFIG_DIR:=$(realpath "$PWD"/../haproxy-config)}"
HAPROXY_CONFIG_DIR=/tmp/${USER:-nouser}-haproxy-config
export HAPROXY_CONFIG_DIR
echo "HAProxy configuration directory: $HAPROXY_CONFIG_DIR"

if [[ ! -f /tmp/haproxy-default.pem ]]; then
    "${thisdir}"/gen-domain-cert.bash
else
    echo -n "reusing existing cert: "; ls -l /tmp/haproxy-default.pem
fi

rm -rf "${HAPROXY_CONFIG_DIR}"
mkdir -p "${HAPROXY_CONFIG_DIR}"/router/{certs,cacerts,whitelists}
mkdir -p "${HAPROXY_CONFIG_DIR}"/{conf/.tmp,run,bin,log}
touch "${HAPROXY_CONFIG_DIR}"/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}
cp "${thisdir}/error-page-404.http" "${HAPROXY_CONFIG_DIR}/conf"
cp "${thisdir}/error-page-503.http" "${HAPROXY_CONFIG_DIR}/conf"

cat <<EOF > "${HAPROXY_CONFIG_DIR}/conf/haproxy.config"
global
  log stdout format raw local0

  maxconn $MAXCONN
  nbthread $NBTHREAD

  # daemon
  ca-base /etc/ssl
  crt-base /etc/ssl
  # TODO: Check if we can get reload to be faster by saving server state.
  # server-state-file /tmp/haproxy.state
  stats socket /tmp/haproxy.sock mode 600 level admin expose-fd listeners
  stats timeout 2m

  # Increase the default request size to be comparable to modern cloud load balancers (ALB: 64kb), affects
  # total memory use when large numbers of connections are open.
  # In OCP 4.8, this value is adjustable via the IngressController API.
  # Cluster administrators are still encouraged to use the default values provided below.
  tune.maxrewrite 8192
  tune.bufsize 32768

  # Configure the TLS versions we support
  ssl-default-bind-options ssl-min-ver TLSv1.2

# The default cipher suite can be selected from the three sets recommended by https://wiki.mozilla.org/Security/Server_Side_TLS,
# or the user can provide one using the ROUTER_CIPHERS environment variable.
# By default when a cipher set is not provided, intermediate is used.
  # user provided list of ciphers (Colon separated list as seen above)
  # the env default is not used here since we can't get here with empty ROUTER_CIPHERS
  tune.ssl.default-dh-param 2048
  ssl-default-bind-ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384

  ssl-default-bind-ciphersuites TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256

defaults
  maxconn $MAXCONN

  # To configure custom default errors, you can either uncomment the
  # line below (server ... 127.0.0.1:8080) and point it to your custom
  # backend service or alternatively, you can send a custom 503 or 404 error.
  #
  # server openshift_backend 127.0.0.1:8080
  errorfile 503 ${HAPROXY_CONFIG_DIR}/conf/error-page-503.http
  errorfile 404 ${HAPROXY_CONFIG_DIR}/conf/error-page-404.http

  timeout connect 5s
  timeout client 30s
  timeout client-fin 1s
  timeout server 30s
  timeout server-fin 1s
  timeout http-request 10s
  timeout http-keep-alive 300s

  # Long timeout for WebSocket connections.
  timeout tunnel 1h



frontend public

  bind :8080
  mode http
  tcp-request inspect-delay 5s
  tcp-request content accept if HTTP
  monitor-uri /_______internal_router_healthz

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]

  # check if we need to redirect/force using https.
  acl secure_redirect base,map_reg_int(${HAPROXY_CONFIG_DIR}/conf/os_route_http_redirect.map) -m bool
  redirect scheme https if secure_redirect

  use_backend %[base,map_reg(${HAPROXY_CONFIG_DIR}/conf/os_http_be.map)]

  default_backend openshift_default

# public ssl accepts all connections and isn't checking certificates yet certificates to use will be
# determined by the next backend in the chain which may be an app backend (passthrough termination) or a backend
# that terminates encryption in this router (edge)
frontend public_ssl

  bind :8443
  tcp-request inspect-delay 5s
  tcp-request content accept if { req_ssl_hello_type 1 }

  # if the connection is SNI and the route is a passthrough don't use the termination backend, just use the tcp backend
  # for the SNI case, we also need to compare it in case-insensitive mode (by converting it to lowercase) as RFC 4343 says
  acl sni req.ssl_sni -m found
  acl sni_passthrough req.ssl_sni,lower,map_reg(${HAPROXY_CONFIG_DIR}/conf/os_sni_passthrough.map) -m found
  use_backend %[req.ssl_sni,lower,map_reg(${HAPROXY_CONFIG_DIR}/conf/os_tcp_be.map)] if sni sni_passthrough

  # if the route is SNI and NOT passthrough enter the termination flow
  use_backend be_sni if sni

  # non SNI requests should enter a default termination backend rather than the custom cert SNI backend since it
  # will not be able to match a cert to an SNI host
  default_backend be_no_sni

##########################################################################
# TLS SNI
#
# When using SNI we can terminate encryption with custom certificates.
# Certs will be stored in a directory and will be matched with the SNI host header
# which must exist in the CN of the certificate.  Certificates must be concatenated
# as a single file (handled by the plugin writer) per the haproxy documentation.
#
# Finally, check re-encryption settings and re-encrypt or just pass along the unencrypted
# traffic
##########################################################################
backend be_sni
  server fe_sni unix@/tmp/haproxy-sni.sock weight 1 send-proxy

frontend fe_sni
  # log global
  # option httplog
  # option dontlognull

  # terminate ssl on edge
  bind unix@/tmp/haproxy-sni.sock ssl crt /tmp/haproxy-default.pem crt-list ${HAPROXY_CONFIG_DIR}/conf/cert_config.map accept-proxy
  mode http

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]



  # map to backend
  # Search from most specific to general path (host case).
  # Note: If no match, haproxy uses the default_backend, no other
  #       use_backend directives below this will be processed.
  use_backend %[base,map_reg(${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map)]

  default_backend openshift_default

##########################################################################
# END TLS SNI
##########################################################################

##########################################################################
# TLS NO SNI
#
# When we don't have SNI the only thing we can try to do is terminate the encryption
# using our wild card certificate.  Once that is complete we can either re-encrypt
# the traffic or pass it on to the backends
##########################################################################
# backend for when sni does not exist, or ssl term needs to happen on the edge
backend be_no_sni
  server fe_no_sni unix@/tmp/haproxy-no-sni.sock weight 1 send-proxy

frontend fe_no_sni

  # terminate ssl on edge
  bind unix@/tmp/haproxy-no-sni.sock ssl crt /tmp/haproxy-default.pem accept-proxy
  mode http

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]



  # map to backend
  # Search from most specific to general path (host case).
  # Note: If no match, haproxy uses the default_backend, no other
  #       use_backend directives below this will be processed.
  use_backend %[base,map_reg(${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map)]

  default_backend openshift_default

listen stats
  bind :1936
  log global
  option httplog
  mode http
  stats enable
  stats refresh 5s
  stats uri /stats

##########################################################################
# END TLS NO SNI
##########################################################################

backend openshift_default
  mode http
  option forwardfor
  #option http-keep-alive
  option http-pretend-keepalive

##-------------- app level backends ----------------"
EOF

for name in $(backend_names_sorted); do
    http_port=${BACKEND_HTTP_PORTS[$name]}
    https_port=${BACKEND_HTTPS_PORTS[$name]}
    container_id=${BACKEND_CONTAINER_IDS[$name]}

    echo "
backend be_secure:${name}-reencrypt
  mode http
  option redispatch
  option forwardfor
  balance random

  timeout check 5000ms
  http-request add-header X-Forwarded-Host %[req.hdr(host)]
  http-request add-header X-Forwarded-Port %[dst_port]
  http-request add-header X-Forwarded-Proto http if !{ ssl_fc }
  http-request add-header X-Forwarded-Proto https if { ssl_fc }
  http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }
  http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]
  cookie $(rev <<<"$container_id") insert indirect nocache httponly secure attr SameSite=None
  server pod:${name}:${host_ip}:$https_port ${host_ip}:$https_port cookie $container_id weight 1 ssl verify required ca-file ${HAPROXY_CONFIG_DIR}/router/cacerts/be_secure:${name}-reencrypt.pem

backend be_edge_http:${name}-edge
  mode http
  option redispatch
  option forwardfor
  balance random

  timeout check 5000ms
  http-request add-header X-Forwarded-Host %[req.hdr(host)]
  http-request add-header X-Forwarded-Port %[dst_port]
  http-request add-header X-Forwarded-Proto http if !{ ssl_fc }
  http-request add-header X-Forwarded-Proto https if { ssl_fc }
  http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }
  http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]
  cookie $(rev <<<"$container_id") insert indirect nocache httponly secure attr SameSite=None
  server pod:${name}:${host_ip}:$http_port ${host_ip}:$http_port cookie $container_id weight 1

backend be_http:${name}-http
  mode http
  option redispatch
  option forwardfor
  balance random

  timeout check 5000ms
  http-request add-header X-Forwarded-Host %[req.hdr(host)]
  http-request add-header X-Forwarded-Port %[dst_port]
  http-request add-header X-Forwarded-Proto http if !{ ssl_fc }
  http-request add-header X-Forwarded-Proto https if { ssl_fc }
  http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }
  http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]
  cookie $(rev <<<"$container_id") insert indirect nocache httponly secure attr SameSite=None
  server pod:${name}:${host_ip}:$http_port ${host_ip}:$http_port cookie $container_id weight 1

backend be_tcp:${name}-passthrough
  balance source

  hash-type consistent
  timeout check 5000ms
  server pod:${name}:${host_ip}:$https_port ${host_ip}:$https_port weight 1" >> "${HAPROXY_CONFIG_DIR}/conf/haproxy.config"
done

for name in $(backend_names_sorted); do
    echo "^${name}-edge.${domain}\.?(:[0-9]+)?(/.*)?$ be_edge_http:${name}-edge" >> "${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map"
    echo "^${name}-reencrypt.${domain}\.?(:[0-9]+)?(/.*)?$ be_secure:${name}-reencrypt" >> "${HAPROXY_CONFIG_DIR}/conf/os_edge_reencrypt_be.map"
done

for name in $(backend_names_sorted); do
    echo "^${name}-http.${domain}\.?(:[0-9]+)?(/.*)?$ be_http:${name}-http" >> "${HAPROXY_CONFIG_DIR}/conf/os_http_be.map"
done

for name in $(backend_names_sorted); do
    echo "^${name}-passthrough.${domain}$ 1" >> "${HAPROXY_CONFIG_DIR}/conf/os_sni_passthrough.map"
done

for name in $(backend_names_sorted); do
    echo "${name}-passthrough.${domain}\.?(:[0-9]+)?(/.*)?$ be_tcp:${name}-passthrough" >> "${HAPROXY_CONFIG_DIR}/conf/os_tcp_be.map"
    echo "${name}-reencrypt.${domain}\.?(:[0-9]+)?(/.*)?$ be_secure:${name}-reencrypt" >> "${HAPROXY_CONFIG_DIR}/conf/os_tcp_be.map"
done

dest_cacrt="$(cat "${thisdir}"/certs/reencrypt/destca.crt)"
tls_crt="$(cat "${thisdir}"/certs/reencrypt/tls.crt)"
tls_key="$(cat "${thisdir}"/certs/reencrypt/tls.key)"

for name in $(backend_names_sorted); do
    printf "%s\n" "${dest_cacrt}" > "${HAPROXY_CONFIG_DIR}/router/cacerts/be_secure:${name}-reencrypt.pem"
    printf "%s\n%s\n" "$tls_key" "$tls_crt" > "${HAPROXY_CONFIG_DIR}/router/certs/be_secure:${name}-reencrypt.pem"
    echo "${HAPROXY_CONFIG_DIR}/router/certs/be_secure:${name}-reencrypt.pem [alpn h2,http/1.1] ${name}-reencrypt.${domain}" >> "${HAPROXY_CONFIG_DIR}/conf/cert_config.map"
done