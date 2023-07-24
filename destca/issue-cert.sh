#!/usr/bin/env bash

set -eux
set -o pipefail

domain=""
force_renewal=""
operation="--renew"

PARAMS=""
while (( "$#" )); do
    case "$1" in
	-i|--issue)
	    operation="--issue"
	    shift
	    ;;
	-d|--domain)
	    if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
		domain=$2
		shift 2
	    else
		echo "Error: Argument for $1 is missing" >&2
		exit 1
	    fi
	    ;;
	-f|--force)
	    force_renewal="--force"
	    shift
	    ;;
	--*)
	    echo "Error: Unsupported flag $1" >&2
	    exit 1
	    ;;
	-*)
	    echo "Error: Unsupported flag $1" >&2
	    exit 1
	    ;;
	*) # preserve positional arguments
	    PARAMS="$PARAMS $1"
	    shift
	    ;;
    esac
done

if [[ -z "$domain" ]]; then
    echo "Please supply a domain."
    exit 1
fi

# reset positional arguments
eval set -- "$PARAMS"

export certdir=$PWD/certs/$domain
mkdir -p "$certdir"

"${HOME}/acme.sh/acme.sh" $force_renewal $operation --keylength 2048 --server letsencrypt -d "$domain" --dns dns_cf
"${HOME}/acme.sh/acme.sh" --install-cert -d "$domain" --keylength 2048 --cert-file "$certdir/cert.pem" --key-file "$certdir/key.pem" --fullchain-file "$certdir/fullchain.pem" --ca-file "$certdir/ca.cer"
