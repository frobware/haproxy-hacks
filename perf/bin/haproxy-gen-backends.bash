#!/usr/bin/env bash

set -eu

. common.sh

host_ip=$(dig +search +short $(hostname))

for name in $(docker_pods | sort -V); do
    port="$(docker inspect --format='{{(index (index .NetworkSettings.Ports "8443/tcp") 0).HostPort}}' "$name")"
    container_id="$(docker inspect --format='{{.Id}}' "$name")"
    name=${name//_/-}
    echo "
backend be_secure:${name}
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
  server pod:${name}:https:${host_ip}:$port ${host_ip}:$port cookie $container_id weight 1 ssl verify required ca-file ${HAPROXY_CONFIG_DIR}/router/cacerts/be_secure:${name}.pem"
done
