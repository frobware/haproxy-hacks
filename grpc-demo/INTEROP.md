# Run gRPC interop tests on OpenShift cluster

### Ensure http/2 is enabled

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=true

Wait for new router pods to rollout.

### Add gRPC test server and verify

    $ oc process -f router-grpc-interop.yaml | oc apply -f -

    $ oc get pods
    NAME                     READY   STATUS    RESTARTS   AGE
    grpc-interop             1/1     Running   0          82m

### Add gRPC routes and verify

    $ ./grpc-interop-apply-routes.sh

    $ oc get routes
    NAME                       HOST/PORT                                               PATH   SERVICES       PORT    TERMINATION            WILDCARD
    grpc-interop-edge          grpc-interop-edge.apps.ocp410.int.frobware.com                 grpc-interop   1110    edge/Redirect          None
    grpc-interop-h2c           grpc-interop-h2c.apps.ocp410.int.frobware.com                  grpc-interop   1110                           None
    grpc-interop-passthrough   grpc-interop-passthrough.apps.ocp410.int.frobware.com          grpc-interop   8443    passthrough/Redirect   None
    grpc-interop-reencrypt     grpc-interop-reencrypt.apps.ocp410.int.frobware.com            grpc-interop   8443    reencrypt/Redirect     None

### Run gRPC tests against grpc-interop routes

    $ go run grpc-client-test.go -domain $(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
