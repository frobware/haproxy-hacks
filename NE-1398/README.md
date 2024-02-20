# Enable EIPs on an OpenShift IngressController

https://issues.redhat.com/browse/NE-1398

This is a demo of how to enable AWS EIPs on an OpenShift AWS/NLB
IngressController.

## Install/Clone

    $ git clone https://github.com/frobware/haproxy-hacks
    $ cd NE-1398

## Find your AWS VPC_ID

    $ VPC_ID=$(./discover-vpcid.sh)
    $ echo $VPC_ID
    vpc-<redacted>

## List the public subnets:

    $ ./list-public-subnets.sh $VPC_ID
    subnet-<redacted>
    subnet-<redacted>
    subnet-<redacted>
    subnet-<redacted>
    subnet-<redacted>

## List unassociated EIPs

    $ ./list-unassociated-eips.sh

The likelihood is nothing is listed.

## Create 5 EIPs (as we have 5 public subnets)

The
[documentation](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.4/guide/service/annotations/#eip-allocations)
explicitly requires that the number of Elastic IPs created must match
the number of public-facing subnets. This is highlighted under the
section for EIP allocations, where it states, "Length must match the
number of subnets."

```sh
$ ./create-unassociated-eips.sh 5
EIP allocated successfully. Allocation ID: eipalloc-<redacted>
EIP allocated successfully. Allocation ID: eipalloc-<redacted>
EIP allocated successfully. Allocation ID: eipalloc-<redacted>
EIP allocated successfully. Allocation ID: eipalloc-<redacted>
EIP allocated successfully. Allocation ID: eipalloc-<redacted>
```

When using the AWS console to inspect each of these new EIP IDs,
you'll observe that they are not associated with any Elastic Network
Interface (ENI). The are "unassociated".

To facilitate a future step, let's store the list of EIP allocation
IDs in a shell variable:

```sh
$ EIPS=$(./eip-allocations-string.sh)
$ echo $EIPS
eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>
```

## Custom build of the cluster-ingress operator

We need a custom version of the cluster ingress-operator to tie things
together:

    $ git clone https://github.com/frobware/cluster-ingress-operator.git -b ne1398-enable-eip-allocations
    $ cd ~/cluster-ingress-operator
    $ make

Elastic IPs (EIPs) can be assigned to a Service with the type
`LoadBalancer` by applying the annotation
`service.beta.kubernetes.io/aws-load-balancer-eip-allocations:
<COMMA-SEPARATED_EIP-ALLOCATIONS>`.

My branch/hack, named `ne1398-enable-eip-allocations`, introduces
functionality to add the EIP allocation annotation if a corresponding
environment variable is set. For instance, if I create an ingress
controller named "ne1398", I would then run a local instance of the
ingress-operator like this:

```sh
$ ne1398_EIP_ALLOCATIONS="$EIPS" ./hack/run.local.sh
```

In this setup, the `ne1398_EIP_ALLOCATIONS` environment variable is
used to specify the EIP allocations, which are then applied as
annotations to the "ne1398" ingress controller. We've used the
existing shell variable (`$EIPS`) from our previous step.

It's critical to understand that EIPs can be associated only during
the creation phase. If we navigate back to the `haproxy-hacks/NE-1398`
directory we can create the ingress controller with the following
helper script:

```sh
$ cd ~/haproxy-hacks/NE-1398
$ ./create-ingresscontroller-nlb.sh ne1398
ingresscontroller.operator.openshift.io/ne1398 created
```

If you look at the output from the `ingress-operator` you will see
confirmation that the environment variable was detected.

The ingress-operator logs will show:

```console
2024-02-21T14:56:33.403Z        INFO    operator.ingress_controller     ingress/load_balancer_service.go:744    adding EIP allocations  {"openshift-ingress-operator": "namespace", "ne1398": "name", "eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>,eipalloc-<redacted>": "eipAllocations"}
```

If, after a bit of "settling time", you revisit the AWS console and
look at each EIP allocation you'll now see that they are attached to a
Elastic Network Interface (i.e., which is ultimately the router).

## CLEANING UP

Once you've finished with the demo don't forget to delete the
allocated EIPs after the ingress controller has been deleted.

```sh
$ oc delete ingresscontrollers  -n openshift-ingress-operator ne1398
ingresscontroller.operator.openshift.io "ne1398" deleted

$ ./release-unassociated-eips.sh
Released EIP 107.20.178.134 with Allocation ID: eipalloc-<redacted>
Released EIP 18.204.101.52 with Allocation ID: eipalloc-<redacted>
Released EIP 34.235.253.182 with Allocation ID: eipalloc-<redacted>
Released EIP 52.204.111.225 with Allocation ID: eipalloc-<redacted>
Released EIP 52.73.177.96 with Allocation ID: eipalloc-<redacted>
All unassociated EIPs have been released.
```
