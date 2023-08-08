#!/usr/bin/env bash

set -eu

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
publicblog_certdir="$PWD/certs/medicalrecords-${project}.${domain}"

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
	    medicalrecords_certdir=~/src/github.com/frobware/infra/ocp414.int.frobware.com/ingress-certs
	    publicblog_certdir=~/src/github.com/frobware/infra/ocp414.int.frobware.com/ingress-certs
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

pushd ocpbugs12858-test
oc apply -f service.yaml
oc apply -f deployment.yaml
./process-routes.sh "$publicblog_certdir" publicblog-route.yaml
./process-routes.sh "$medicalrecords_certdir" medicalrecords-route.yaml
popd

oc rollout status "deployment/${project}-test"
oc get all
