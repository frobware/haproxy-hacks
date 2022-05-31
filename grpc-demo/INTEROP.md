# Run gRPC interop tests

### Ensure http/2 is enabled

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=true

## Wait for new router pods to rollout.

## Add gRPC test server

    $ router-grpc-interop.yaml

## Add gRPC routes

    $ ./grpc-interop-apply-routes.sh

## Run gRPC tests against grpc-interop routes

    $ go run grpc-client-test.go -domain $(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
