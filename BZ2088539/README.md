# Setup

Reproducer for: https://bugzilla.redhat.com/show_bug.cgi?id=2088539

Run only 1 replica (for greppability), and enable access logging:

    $ oc -n openshift-ingress-operator patch ingresscontroller/default --type=merge --patch='{"spec":{"logging":{"access":{"destination":{"type":"Container"}}}}}'
    $ oc -n openshift-ingress-operator scale --replicas=1 ingresscontroller/default

Wait for new router pods to rollout.

# Deploy a hello openshift app

    $ oc apply -f deployment.yaml

# Create route with its own certificate (for when we we enable HTTP/2)

    $ ./create-routes.sh

# Reproducer steps

## Disable HTTP/2 on the `default` ingresscontroller (this is the default)

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=false
    # Wait for the router deployment to rollout...

### Prove the app works without HTTP/2 enabled on the `default` ingresscontroller

    $ curl -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

### Prove the app works with double-slash in the URL with HTTP/1.1 and HTTP/2

    $ curl --http1.1 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

    $ curl --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    Hello OpenShift!

## Enable HTTP/2 on the `default` ingresscontroller

    $ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=true

    # Wait for the deployment to rollout...

    $ oc get pods -n openshift-ingress -w
    NAME                              READY   STATUS    RESTARTS   AGE
    router-default-795bbc5887-j9829   2/2     Running   0          4m32s
    router-default-8678f494b9-js2zm   1/2     Running   0          14s
    router-default-8678f494b9-js2zm   2/2     Running   0          21s
    router-default-795bbc5887-j9829   2/2     Terminating   0        4m39s

### Prove the app works when there are no double-slashes in the URL

    $ curl --http1.1 -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

    $ curl --http2 -k -L hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)
    Hello OpenShift!

### Prove HAProxy routing is broken with double-slash

- OCP 4.11 OK
- OCP 4.10 OK
- OCP 4.9  NOT OK
- OCP 4.8  NOT OK
- OCP 4.7  OK
- OCP 4.6  OK

    $ curl --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)

    $ curl --http2-prior-knowledge -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)

    $ curl --http2 -k -L https://hello-edge.$(oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)//foo/bar/baz
    curl: (92) HTTP/2 stream 0 was not closed cleanly: PROTOCOL_ERROR (err 1)
