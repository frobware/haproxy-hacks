#!/usr/bin/env bash

set -eu

: "${ROUTER_CERT_NAME:=router-certs-default}"

project="$(oc project -q)"

if [[ "$project" != "ocpbugs12858" ]]; then
    echo "Expecting current namespace to \"ocpbugs12858\"."
    exit 1
fi

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

medicalrecords_certdir="$PWD/certs/medicalrecords-${project}.${domain}"
publicblog_certdir="$PWD/certs/publicblog-${project}.${domain}"
use_wildcard_domain=0
wildcard_cert_dir=$(mktemp -d)
add_destca=0

trap 'rm -rf -- "$wildcard_cert_dir"' EXIT INT TERM HUP QUIT

PARAMS=""
while (( "$#" )); do
    case "$1" in
	--medicalrecords_certdir)
	    if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
		medicalrecords_certdir=$2
		shift 2
	    else
		echo "Error: Argument for $1 is missing" >&2
		exit 1
	    fi
	    ;;
	--publicblog_certdir)
	    if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
		publicblog_certdir=$2
		shift 2
	    else
		echo "Error: Argument for $1 is missing" >&2
		exit 1
	    fi
	    ;;
	-w|--use_wildcard_domain)
	    use_wildcard_domain=1
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

dest_cacrt="$(oc exec -n openshift-ingress -c router "$pod" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt)"

oc delete --all routes
oc delete --all services
oc delete --all deployments

make -C server

pushd browser-test
oc apply -f deployment.yaml
oc apply -f service.yaml
oc apply -f routes.yaml
popd

if [[ $use_wildcard_domain -eq 1 ]]; then
    echo "Extracting secrets from $ROUTER_CERT_NAME"
    oc extract secret/$ROUTER_CERT_NAME -n openshift-ingress --keys=tls.crt --to=- > "$wildcard_cert_dir/tls.crt"
    oc extract secret/$ROUTER_CERT_NAME -n openshift-ingress --keys=tls.key --to=- > "$wildcard_cert_dir/tls.key"
    medicalrecords_certdir="$wildcard_cert_dir"
    publicblog_certdir="$wildcard_cert_dir"
    ls -lR $wildcard_cert_dir
fi

pushd ocpbugs12858-test
oc apply -f service.yaml
oc apply -f deployment.yaml

echo -n "medicalrecords CN: "
openssl x509 -in $medicalrecords_certdir/tls.crt -noout -subject | awk -F= '/CN/ {print $NF}'

echo -n "publicblog_certdir CN: "
openssl x509 -in $publicblog_certdir/tls.crt -noout -subject | awk -F= '/CN/ {print $NF}'

./process-routes.sh "$publicblog_certdir" publicblog-route.yaml
./process-routes.sh "$medicalrecords_certdir" medicalrecords-route.yaml
popd

oc rollout status "deployment/${project}-test"
oc get all
