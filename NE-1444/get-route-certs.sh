#!/usr/bin/env bash

extract_certificate() {
    echo | openssl s_client -connect "$1":443 -showcerts 2>/dev/null | sed -n '/BEGIN CERTIFICATE/,/END CERTIFICATE/p'
}

hostname1="catpictures-ne1444.apps.ocp416.int.frobware.com"
hostname2="payroll-ne1444.apps.ocp416.int.frobware.com"

echo "Certificate for $hostname1:"
extract_certificate "$hostname1"
echo
echo "Certificate for $hostname2:"
extract_certificate "$hostname2"
