# Setup

Reproducer for: https://bugzilla.redhat.com/show_bug.cgi?id=2088539

Run only 1 replica (for greppability), and enable access logging:

    $ oc -n openshift-ingress-operator patch ingresscontroller/default --type=merge --patch='{"spec":{"logging":{"access":{"destination":{"type":"Container"}}}}}'
    $ oc -n openshift-ingress-operator scale --replicas=1 ingresscontroller/default

Wait for new router pods to rollout.

# Deploy a hello openshift app

    $ oc apply -f deployment.yaml

# Create route with its own certificate (for when we we enable HTTP/2)

    $ go run certgen/certgen.go > /tmp/env
    $ . /tmp/env
    $ oc process -p TLS_CRT="$TLS_CRT" -p TLS_KEY="$TLS_KEY" -p DOMAIN="$domain" -f router-grpc-interop-routes.yaml | oc apply -f -

# Reproducer steps

## Disable HTTP/2 on the `default` ingresscontroller (this is the default)

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=false
    # Wait for the router deployment to rollout...

### Prove the app works without HTTP/2 enabled on the `default` ingresscontroller

    $ curl -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

### Prove the app works with double-slash in the URL with HTTP/1.1

    $ curl --http1.1 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

### Prove the app works with double-slash in the URL with HTTP/2

    $ curl --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

    $ curl --http2-prior-knowledge -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

    $ curl --http2-prior-knowledge --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

## Enable HTTP/2 on the `default` ingresscontroller

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=true
    # Wait for the deployment to rollout...

### Prove the app works when there are no double-slashes in the URL

    $ curl --http1.1 -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

    $ curl --http2 -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

### Prove HAProxy routing is broken with double-slash

    $ curl --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)

    $ curl --http2-prior-knowledge -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)

    $ curl --http2-prior-knowledge --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)

# Summary

    OCP 4.8.25  is broken (haproxy22-2.2.13-3.el8.x86_64)

    OCP 4.10.0 and onwards is OK (upstream fix is in haproxy-2.2.17)
    OCP 4.11.0 and onwards is OK (upstream fix is in haproxy-2.2.24)
