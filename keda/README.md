# KEDA PoC/Demo for openshift-ingress

## Useful links

- https://github.com/zroubalik/keda-openshift-examples/tree/main/prometheus
- https://access.redhat.com/articles/6718611
- https://issues.redhat.com/browse/RHSTOR-1938

# Setup

## Install KEDA

Install KEDA operator (v2.7.1); I used the OperatorHub from within the
console. Don't forget to create a `KedaController` (the operator hints
at this once installed).

Verify that KEDA is installed:

```console
$ oc get pods -n keda
NAME                                     READY   STATUS    RESTARTS   AGE
keda-metrics-apiserver-f57cbb4b5-2gssq   1/1     Running   0          36s
keda-olm-operator-79fd785d5d-s4zms       1/1     Running   0          2m10s
keda-operator-78b964d4cd-7rclt           1/1     Running   0          36s
```

# Goal/Demo

Scale an ingresscontroler on some user-defined metric. For example:

- some arbitrary metric (e.g., match replicas from another deployment)
- the number of Ready and Schedulable worker nodes

## Setup

We're going to do everything in the `openshift-ingress-operator`
namespace.

    $ oc project openshift-ingress-operator
    $ git clone https://github.com/frobware/haproxy-hacks && cd keda

## Create KEDA pieces for autoscaling

This [existing
demo](https://github.com/zroubalik/keda-openshift-examples/tree/main/prometheus/ocp-monitoring)
details a number of steps that need to be run which I have
encapsulated.

    $ ./setup/setup.sh

## Scaling on the number of replicas from an existing deployment.

This is a very arbitrary demo but I wanted to connect the dots and
gain some understanding. Deploy the `hello-app` and `test-app`.

    $ oc create -f ./test-app/deployment.yaml
    $ oc create -f ./hello-app/deployment.yaml

## Our first scaledobject


Run the [setup/setup.sh] script
