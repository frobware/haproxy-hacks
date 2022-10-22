#!/usr/bin/env bash

set -eu

if [[ -z "$(hostname -d)" ]]; then
    domain="localdomain"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "$tmpdir"' EXIT

openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=MyCompany/CN=$(hostname -s).$(hostname -d)" -keyout "$tmpdir/ca.key" -out "$tmpdir/ca.crt"
openssl req -out "$tmpdir/edge.csr" -newkey rsa:2048 -nodes -keyout "$tmpdir/edge.key" -subj "/CN=*.$(hostname -d)/O=MyCompany"
openssl x509 -req -sha256 -days 365 -CA "$tmpdir/ca.crt" -CAkey "$tmpdir/ca.key" -set_serial 0 -in "$tmpdir/edge.csr" -out "$tmpdir/edge.crt"
cat "$tmpdir/edge.crt" "$tmpdir/ca.crt" "$tmpdir/edge.key" > /tmp/haproxy-default.pem
echo "haproxy SSL certificate finalised in /tmp/haproxy-default.pem"
