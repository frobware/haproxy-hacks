#!/usr/bin/env bash

set -eu

# domain="int.frobware.com"
# host=$(dig +search +short $(hostname))

[[ -d "${1:?}/conf" ]] || {
    echo "$1/conf directory not found";
    exit 1
}

# this is from: https://github.com/openshift-scale/images/tree/master/nginx
nginx_dest_cacrt="-----BEGIN CERTIFICATE-----
MIIDbTCCAlWgAwIBAgIJAJR/jN0Oa+/rMA0GCSqGSIb3DQEBCwUAME0xCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMQswCQYDVQQHDAJOWTEcMBoGA1UE
CgwTRGVmYXVsdCBDb21wYW55IEx0ZDAeFw0xNzAxMjQwODExMDJaFw0yNzAxMjIw
ODExMDJaME0xCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMQswCQYD
VQQHDAJOWTEcMBoGA1UECgwTRGVmYXVsdCBDb21wYW55IEx0ZDCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBAMItGS9sSafyqBuOcQcQ5j7OQ0EwF9qOckhl
fT8VzUbcOy8/L/w654MpLEa4O4Fiek3keE7SDWGVtGZWDvT9y1QUxPhkDWq1Y3rr
yMelv1xRIyPVD7EEicga50flKe8CKd1U3D6iDQzq0uxZZ6I/VArXW/BZ4LfPauzN
9EpCYyKq0fY7WRFIGouO9Wu800nxcHptzhLAgSpO97aaZ+V+jeM7n7fchRSNrpIR
zPBl/lIBgCPJgkax0tcm4EIKIwlG+jXWc5mvV8sbT8rAv32HVuaP6NafyWXXP3H1
oBf2CQCcwuM0sM9ZeZ5JEDF/7x3eNtqSt1X9HjzVpQjiVBXY+E0CAwEAAaNQME4w
HQYDVR0OBBYEFOXxMHAA1qaKWlP+gx8tKO2rQ81WMB8GA1UdIwQYMBaAFOXxMHAA
1qaKWlP+gx8tKO2rQ81WMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AJAri7Pd0eSY/rvIIvAvjhDPvKt6gI5hJEUp+M3nWTWA/IhQFYutb9kkZGhbBeLj
qneJa6XYKaCcUx6/N6Vvr3AFqVsbbubbejRpdpXldJC33QkwaWtTumudejxSon24
W/ANN/3ILNJVMouspLRGkFfOYp3lq0oKAlNZ5G3YKsG0znAfqhAVtqCTG9RU24Or
xzkEaCw8IY5N4wbjCS9FPLm7zpzdg/M3A/f/vrIoGdns62hzjzcp0QVTiWku74M8
v7/XlUYYvXOvPQCCHgVjnAZlnjcxMTBbwtdwfxjAmdNTmFFpASnf0s3b287zQwVd
IeSydalVtLm7rBRZ59/2DYo=
-----END CERTIFICATE-----
"

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "$tmpdir"' EXIT

go build -o "$tmpdir/certgen" ./certgen/certgen.go

for name in $(docker ps --no-trunc --filter name=^/docker_nginx_ --format '{{.Names}}' | sort -V); do
    name=$(echo $name | sed 's/_/-/g')
    "$tmpdir/certgen" > "$tmpdir/env"
    . "$tmpdir/env"
    # printf "%s\n" "$TLS_CACRT" > "$1/router/cacerts/be_secure:${name}.pem"
    printf "%s\n" "$nginx_dest_cacrt" > "$1/router/cacerts/be_secure:${name}.pem"
    printf "%s\n%s\n%s\n" "$TLS_KEY" "$TLS_CRT" "$TLS_CACRT" > "$1/router/certs/be_secure:${name}.pem"
    echo "$1/router/certs/be_secure:${name}.pem ${name}.int.frobware.com" >> "$1/conf/cert_config.map"
done

exit 0
