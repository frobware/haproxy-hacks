#!/usr/bin/env bash

. common.sh

echo $(domain)
set -eu

if [[ -z "$(domain)" ]]; then
    echo "no domain"
    exit 1
fi

mkdir -p certs
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=MyCompany/CN=$(hostname).$(domain)' -keyout certs/ca.key -out certs/ca.crt
openssl req -out certs/edge.csr -newkey rsa:2048 -nodes -keyout certs/edge.key -subj "/CN=*.$(domain)/O=MyCompany"
openssl x509 -req -sha256 -days 365 -CA certs/ca.crt -CAkey certs/ca.key -set_serial 0 -in certs/edge.csr -out certs/edge.crt
cat certs/edge.crt certs/ca.crt certs/edge.key > certs/default.pem
