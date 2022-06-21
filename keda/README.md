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

## Setup Preliminaries

1. Clone this repo

    $ git clone https://github.com/frobware/haproxy-hacks && cd haproxy-hacks/keda

2. Switch namespaces

    $ oc project openshift-ingress-operator

## Create KEDA pieces for autoscaling

This [existing
demo](https://github.com/zroubalik/keda-openshift-examples/tree/main/prometheus/ocp-monitoring)
details a number of steps that need to be run which I have
encapsulated in the following setup script; run that:

    $ ./setup/setup.sh

# Demo 1 - just scale something using KEDA

This is a very arbitrary demo but I wanted to connect the dots and
gain some understanding. Deploy `hello-app` and `test-app`:

    $ oc create -f ./test-app/deployment.yaml
    $ oc create -f ./hello-app/deployment.yaml

    $ oc get pods
    NAME                                READY   STATUS    RESTARTS      AGE
    hello-app-6bddd4f888-lqvjv          1/1     Running   0             4m56s
    ingress-operator-6f866c8d44-cgs68   2/2     Running   2 (77m ago)   84m
    test-app-979b5897-dbptv             1/1     Running   0             5m2s

We're going to scale the `test-app` up and down based on the number of
replicas in the `hello-app` deployment.

    $ oc create -f test-app/scale-on-hello-app.yaml
    scaledobject.keda.sh/test-app-scale-on-hello-app created

Verify that the scaleobject is created and "Ready":

    $ oc get scaledobjects.keda.sh
    NAME                          SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS              AUTHENTICATION   READY   ACTIVE   FALLBACK   AGE
    test-app-scale-on-hello-app   apps/v1.Deployment   test-app          1     20    kubernetes-workload                    True    True     False      79s

And under the hood we have an associated `HorizontalPodAutoscaler`:

    $ oc get hpa
    NAME                                   REFERENCE             TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
    keda-hpa-test-app-scale-on-hello-app   Deployment/test-app   1/1 (avg)   1         20        1          2m24s

The goal of `scaledobject/test-app-scale-on-hello-app` is to track the
number of replicas in the `hello-app` and scale up/down based on the
number of replicas in that deployment.

    $ oc scale --replicas=10 deployment/hello-app
    deployment.apps/hello-app scaled

    $ oc get pods -l app=hello-app
    NAME                         READY   STATUS    RESTARTS   AGE
    hello-app-6bddd4f888-9fvjr   1/1     Running   0          29s
    hello-app-6bddd4f888-bsj4d   1/1     Running   0          29s
    hello-app-6bddd4f888-h2glp   1/1     Running   0          29s
    hello-app-6bddd4f888-lqvjv   1/1     Running   0          10m
    hello-app-6bddd4f888-ps2lm   1/1     Running   0          29s
    hello-app-6bddd4f888-rdzwg   1/1     Running   0          29s
    hello-app-6bddd4f888-rs8ms   1/1     Running   0          29s
    hello-app-6bddd4f888-vxlgx   1/1     Running   0          29s
    hello-app-6bddd4f888-wjpwm   1/1     Running   0          29s
    hello-app-6bddd4f888-zxcfw   1/1     Running   0          29s

If we look at the HPA we see that we now have a new target of 10
replicas:

    $ oc get hpa
    NAME                                   REFERENCE             TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
    keda-hpa-test-app-scale-on-hello-app   Deployment/test-app   1/1 (avg)   1         20        10         5m12s

and we can verify that the `test-app` has 10 replicas:

    $ oc get pods -l app=test-app
    NAME                      READY   STATUS    RESTARTS   AGE
    test-app-979b5897-9pcnp   1/1     Running   0          116s
    test-app-979b5897-cwcpx   1/1     Running   0          2m26s
    test-app-979b5897-dbptv   1/1     Running   0          12m
    test-app-979b5897-dt4ph   1/1     Running   0          116s
    test-app-979b5897-g2dz2   1/1     Running   0          2m11s
    test-app-979b5897-pt5nr   1/1     Running   0          2m26s
    test-app-979b5897-rdqfn   1/1     Running   0          2m11s
    test-app-979b5897-txxbd   1/1     Running   0          2m26s
    test-app-979b5897-wln5l   1/1     Running   0          2m11s
    test-app-979b5897-xx589   1/1     Running   0          2m11s

Similary scaling down the `hello-app` we see:

    $ oc scale --replicas=0 deployment/hello-app
    deployment.apps/hello-app scaled

**(I have to wait 300s here for reasons I don't currently understand.)**

    $ oc get pods -l app=test-app
    NAME                      READY   STATUS    RESTARTS   AGE
    test-app-979b5897-rdqfn   1/1     Running   0          10m

And the HPA reflects the new target of 1 replica:

    $ oc get hpa
    NAME                                   REFERENCE             TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
    keda-hpa-test-app-scale-on-hello-app   Deployment/test-app   0/1 (avg)   1         20        1          14m

So this is all fine & dandy. I can scale using KEDA. Let's choose
another metric to scale on.

# Demo 2 - scale a deployment on the sum of workers nodes

Let's scale the `test-app` out/in on the number of worker nodes in the
cluster. In my cluster this value is currently 11.

    $ oc get nodes -l  node-role.kubernetes.io/worker | cat -n
    1 NAME                                         STATUS   ROLES    AGE    VERSION
    2 ip-10-0-135-240.us-east-2.compute.internal   Ready    worker   75s    v1.24.0+284d62a
    3 ip-10-0-137-157.us-east-2.compute.internal   Ready    worker   77s    v1.24.0+284d62a
    4 ip-10-0-153-58.us-east-2.compute.internal    Ready    worker   118m   v1.24.0+284d62a
    5 ip-10-0-154-216.us-east-2.compute.internal   Ready    worker   75s    v1.24.0+284d62a
    6 ip-10-0-156-42.us-east-2.compute.internal    Ready    worker   75s    v1.24.0+284d62a
    7 ip-10-0-169-13.us-east-2.compute.internal    Ready    worker   86s    v1.24.0+284d62a
    8 ip-10-0-176-52.us-east-2.compute.internal    Ready    worker   107s   v1.24.0+284d62a
    9 ip-10-0-179-1.us-east-2.compute.internal     Ready    worker   118m   v1.24.0+284d62a
   10 ip-10-0-186-40.us-east-2.compute.internal    Ready    worker   107s   v1.24.0+284d62a
   11 ip-10-0-186-8.us-east-2.compute.internal     Ready    worker   91s    v1.24.0+284d62a
   12 ip-10-0-218-12.us-east-2.compute.internal    Ready    worker   118m   v1.24.0+284d62a

Our metric for this will be: `sum(kube_node_role{role="worker"})`.

    $ oc delete -f test-app/scale-on-hello-app.yaml
    $ oc create -f test-app/scale-on-worker-role.yaml
    scaledobject.keda.sh/test-app-scale-on-worker-role created

    $ oc get scaledobject
    NAME                            SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS              AUTHENTICATION                 READY   ACTIVE   FALLBACK   AGE
    test-app-scale-on-worker-role   apps/v1.Deployment   test-app          1     20    prometheus            keda-trigger-auth-prometheus   True    False    False      36s

    $ oc get hpa
    NAME                                     REFERENCE             TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
    keda-hpa-test-app-scale-on-worker-role   Deployment/test-app   0/1         1         20        1          65s

At this point--and based on our scale out previously with the
`hello-app`--I would expect the `test-app` to scale out but the
`test-app` deployment remains at 1 replica. any scaling.

    $ oc get deployment
    NAME               READY   UP-TO-DATE   AVAILABLE   AGE
    hello-app          0/0     0            0           54m
    ingress-operator   1/1     1            1           134m
    test-app           1/1     1            1           54m

Looking at the scaledobject:

    $ oc get scaledobjects.keda.sh/test-app-scale-on-worker-role -o yaml
    apiVersion: keda.sh/v1alpha1
    kind: ScaledObject
    metadata:
      creationTimestamp: "2022-06-21T10:28:04Z"
      finalizers:
      - finalizer.keda.sh
      generation: 1
      labels:
        scaledobject.keda.sh/name: test-app-scale-on-worker-role
      name: test-app-scale-on-worker-role
      namespace: openshift-ingress-operator
      resourceVersion: "83623"
      uid: b5094f1c-ebce-4439-af4c-2e12d6aeda5d
    spec:
      cooldownPeriod: 5
      maxReplicaCount: 20
      minReplicaCount: 1
      pollingInterval: 2
      scaleTargetRef:
        name: test-app
      triggers:
      - authenticationRef:
          name: keda-trigger-auth-prometheus
        metadata:
          authModes: bearer
          metricName: worker_nodes
          namespace: openshift-ingress-operator
          query: sum(kube_node_role{role="worker"})
          serverAddress: https://thanos-querier.openshift-monitoring.svc.cluster.local:9092
          threshold: "1"
        metricType: Value
        type: prometheus
    status:
      conditions:
      - message: ScaledObject is defined correctly and is ready for scaling
        reason: ScaledObjectReady
        status: "True"
        type: Ready
      - message: Scaling is not performed because triggers are not active
        reason: ScalerNotActive
        status: "False"
        type: Active
      - message: No fallbacks are active on this scaled object
        reason: NoFallbackFound
        status: "False"
        type: Fallback
      externalMetricNames:
      - s0-prometheus-worker_nodes
      health:
        s0-prometheus-worker_nodes:
          numberOfFailures: 0
          status: Happy
      originalReplicaCount: 1
      scaleTargetGVKR:
        group: apps
        kind: Deployment
        resource: deployments
        version: v1
      scaleTargetKind: apps/v1.Deployment

And the HPA:

    $ oc get hpa keda-hpa-test-app-scale-on-worker-role -o yaml
    apiVersion: autoscaling/v2
    kind: HorizontalPodAutoscaler
    metadata:
      creationTimestamp: "2022-06-21T10:28:05Z"
      labels:
        app.kubernetes.io/managed-by: keda-operator
        app.kubernetes.io/name: keda-hpa-test-app-scale-on-worker-role
        app.kubernetes.io/part-of: test-app-scale-on-worker-role
        app.kubernetes.io/version: 2.7.1
        scaledobject.keda.sh/name: test-app-scale-on-worker-role
      name: keda-hpa-test-app-scale-on-worker-role
      namespace: openshift-ingress-operator
      ownerReferences:
      - apiVersion: keda.sh/v1alpha1
        blockOwnerDeletion: true
        controller: true
        kind: ScaledObject
        name: test-app-scale-on-worker-role
        uid: b5094f1c-ebce-4439-af4c-2e12d6aeda5d
      resourceVersion: "86172"
      uid: ddeeef70-ece4-4867-a8be-c9b168f3caba
    spec:
      maxReplicas: 20
      metrics:
      - external:
          metric:
            name: s0-prometheus-worker_nodes
            selector:
              matchLabels:
                scaledobject.keda.sh/name: test-app-scale-on-worker-role
          target:
            type: Value
            value: "1"
        type: External
      minReplicas: 1
      scaleTargetRef:
        apiVersion: apps/v1
        kind: Deployment
        name: test-app
    status:
      conditions:
      - lastTransitionTime: "2022-06-21T10:28:20Z"
        message: recommended size matches current size
        reason: ReadyForNewScale
        status: "True"
        type: AbleToScale
      - lastTransitionTime: "2022-06-21T10:28:20Z"
        message: 'the HPA was able to successfully calculate a replica count from external
          metric s0-prometheus-worker_nodes(&LabelSelector{MatchLabels:map[string]string{scaledobject.keda.sh/name:
          test-app-scale-on-worker-role,},MatchExpressions:[]LabelSelectorRequirement{},})'
        reason: ValidMetricFound
        status: "True"
        type: ScalingActive
      - lastTransitionTime: "2022-06-21T10:33:20Z"
        message: the desired replica count is less than the minimum replica count
        reason: TooFewReplicas
        status: "True"
        type: ScalingLimited
      currentMetrics:
      - external:
          current:
            value: "0"
          metric:
            name: s0-prometheus-worker_nodes
            selector:
              matchLabels:
                scaledobject.keda.sh/name: test-app-scale-on-worker-role
        type: External
      currentReplicas: 1
      desiredReplicas: 1

Notably the value is 0, and not 11.

## Permissions Issues?

Let's deploy another app and try again.

    $ oc create -f nodes-ready-app/deployment.yaml
    deployment.apps/nodes-ready-app created
    service/nodes-ready-app created
    servicemonitor.monitoring.coreos.com/keda-nodes-ready-app-sm created

This app tracks the number of `Ready` and `Schedulable` nodes. Looking
at the results from the pod we see:

    $ oc get pods
    NAME                                READY   STATUS    RESTARTS       AGE
    ingress-operator-6f866c8d44-cgs68   2/2     Running   2 (3h9m ago)   3h16m
    nodes-ready-app-69bf4677bb-x4k9g    1/1     Running   0              2m31s
    test-app-979b5897-rdqfn             1/1     Running   0              106m

    $ oc logs nodes-ready-app-69bf4677bb-x4k9g | head -30
    W0621 11:30:38.562523       1 client_config.go:617] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
    E0621 11:30:39.646476       1 reflector.go:138] k8s.io/client-go/informers/factory.go:134: Failed to watch *v1.Node: unknown (get nodes)
    2022/06/21 11:30:39 Server started on port 8080
    E0621 11:30:41.208452       1 reflector.go:138] k8s.io/client-go/informers/factory.go:134: Failed to watch *v1.Node: unknown (get nodes)
    E0621 11:30:43.518436       1 reflector.go:138] k8s.io/client-go/informers/factory.go:134: Failed to watch *v1.Node: unknown (get nodes)
    E0621 11:30:48.931680       1 reflector.go:138] k8s.io/client-go/informers/factory.go:134: Failed to watch *v1.Node: unknown (get nodes)
    2022/06/21 11:30:49 ip-10-0-154-216.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-156-42.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-186-40.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-135-240.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-137-157.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-169-13.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-153-58.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-220-29.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-164-35.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-186-8.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-176-52.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-179-1.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-218-12.us-east-2.compute.internal READY
    2022/06/21 11:30:49 ip-10-0-154-192.us-east-2.compute.internal READY
    2022/06/21 11:30:49 14 ready nodes
    2022/06/21 11:30:49 0 not ready nodes

And verifying metrics are registered for `ready_nodes`:

    $ oc rsh nodes-ready-app-69bf4677bb-x4k9g curl http://localhost:8080/metrics |grep ready_nodes
    # HELP ready_nodes Report the number of Ready nodes in the cluster.
    # TYPE ready_nodes gauge
    ready_nodes 14

Create our scaledobject:

    $ oc create -f test-app/scale-on-ready-nodes.yaml
    scaledobject.keda.sh/test-app-scale-on-ready-nodes created

    $ oc get scaledobject
    NAME                                                 SCALETARGETKIND      SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION                 READY   ACTIVE   FALLBACK   AGE
    scaledobject.keda.sh/test-app-scale-on-ready-nodes   apps/v1.Deployment   test-app          1     20    prometheus   keda-trigger-auth-prometheus   True    True     False      35s

    $ oc get hpa
    NAME                                                                         REFERENCE             TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
    horizontalpodautoscaler.autoscaling/keda-hpa-test-app-scale-on-ready-nodes   Deployment/test-app   28/1      1         20        8          35s

    $ oc get pods -l app=test-app
    NAME                      READY   STATUS    RESTARTS   AGE
    test-app-979b5897-8hg6v   1/1     Running   0          57s
    test-app-979b5897-9q56k   1/1     Running   0          72s
    test-app-979b5897-b7gdd   1/1     Running   0          102s
    test-app-979b5897-bf7w9   1/1     Running   0          72s
    test-app-979b5897-c5vf4   1/1     Running   0          72s
    test-app-979b5897-dcbtz   1/1     Running   0          102s
    test-app-979b5897-jn8fd   1/1     Running   0          87s
    test-app-979b5897-kgjxb   1/1     Running   0          72s
    test-app-979b5897-mlcm9   1/1     Running   0          72s
    test-app-979b5897-p4qgz   1/1     Running   0          87s
    test-app-979b5897-p6689   1/1     Running   0          72s
    test-app-979b5897-p8vpb   1/1     Running   0          57s
    test-app-979b5897-pl9pl   1/1     Running   0          72s
    test-app-979b5897-qn9qg   1/1     Running   0          87s
    test-app-979b5897-qpgcz   1/1     Running   0          57s
    test-app-979b5897-qxppz   1/1     Running   0          72s
    test-app-979b5897-rdqfn   1/1     Running   0          117m
    test-app-979b5897-sxd8j   1/1     Running   0          87s
    test-app-979b5897-tfctv   1/1     Running   0          102s
    test-app-979b5897-wx8f8   1/1     Running   0          57s

So that worked.
