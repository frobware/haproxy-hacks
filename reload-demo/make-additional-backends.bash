#!/usr/bin/bash

for i in $(seq 3001 13000); do
    cat <<EOF
backend be_http:default:helloworld-${i}
  mode http
  option redispatch
  option forwardfor
  balance leastconn

  timeout check 5000ms
  http-request set-header X-Forwarded-Host %[req.hdr(host)]
  http-request set-header X-Forwarded-Port %[dst_port]
  http-request set-header X-Forwarded-Proto http if !{ ssl_fc }
  http-request set-header X-Forwarded-Proto https if { ssl_fc }
  http-request set-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }
  # Forwarded header: quote IPv6 addresses and values that may be empty as per https://tools.ietf.org/html/rfc7239
  http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)];proto-version=\"%[req.hdr(X-Forwarded-Proto-Version)]\"
  cookie a5313a2bcc9e845108addb4bdf97a72d insert indirect nocache httponly
EOF
done
