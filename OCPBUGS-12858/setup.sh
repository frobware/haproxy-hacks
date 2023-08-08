#!/usr/bin/env bash

set -eu

if [[ "$(oc project -q)" != "ocpbugs12858" ]]; then
    echo "Expecting current namespace to \"ocpbugs12858\"."
    exit 1
fi

domain=$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

tls_key="$(cat /home/aim/.acme.sh/*.${domain}/*.${domain}.key)"
tls_crt="$(cat /home/aim/.acme.sh/*.${domain}/*.${domain}.cer)"
tls_cacrt="$(cat /home/aim/.acme.sh/*.${domain}/ca.cer)"

# Find the first router pod in the openshift-ingress namespace.
pod=$(oc get pods -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -o jsonpath='{.items[0].metadata.name}')

if [[ -z "$pod" ]]; then
    echo "No router pod found in the openshift-ingress namespace."
    exit 1
fi

dest_cacrt="$(oc exec -n openshift-ingress -c router "$pod" -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt)"

oc delete --all routes
oc delete --all services
oc delete --all deployments

make -C server

pushd browser-test
oc apply -f deployment.yaml
oc apply -f service.yaml
oc process TLS_KEY="$tls_key" TLS_CRT="$tls_crt" TLS_CACRT="$tls_cacrt" -f ./routes.yaml -o yaml | oc delete --ignore-not-found -f -
oc process TLS_KEY="$tls_key" TLS_CRT="$tls_crt" TLS_CACRT="$tls_cacrt" -f ./routes.yaml -o yaml | oc apply -f -
popd

pushd ocpbugs12858-test
oc apply -f service.yaml
oc apply -f deployment.yaml
./process-routes.sh ../certs/publicblog-ocpbugs12858.apps.ocp414.int.frobware.com publicblog-route.yaml
./process-routes.sh ../certs/medicalrecords-ocpbugs12858.apps.ocp414.int.frobware.com medicalrecords-route.yaml
popd

#oc get all
