# OpenShift Ingress: testing websocket behaviour

This was tested on:

```
$ oc version
Client Version: 4.10.27
Server Version: 4.11.2
Kubernetes Version: v1.24.0+b62823b
```

This repo contains a "websocket server" deployment that allows you to
test and verify websocket behaviour with the following OpenShift
Ingress route types:

- edge
- insecure
- reencrypt

# Optionally Build server image

This is optional.

I have pushed an image to `quay.io/amcdermo/ocpbugs-173-server`. Use
the `latest` tag.

## Prep environment

```
$ cat <<EOF > .envrc
set -a
REGISTRY=quay.io
IMAGE_ORG=amcdermo
set +a
EOF

$ direnv allow
```

## Build websocket server image

```
$ make -C server push-image
make: Entering directory '/home/aim/src/github.com/frobware/haproxy-hacks/OCPBUGS-173/server'
go fmt ./...
go vet ./...
go mod vendor
go mod tidy
CGO_ENABLED=0 go build -mod=vendor -o server .
rm -f server
podman build -f Containerfile -t amcdermo/ocpbugs-173-server:latest
[1/2] STEP 1/5: FROM registry.access.redhat.com/ubi8/go-toolset AS builder
[1/2] STEP 2/5: USER root
--> Using cache 6aa190c154265b4bfdf08a31ef619bf3963b0859495779f9bd0f1272fd06420b
--> 6aa190c1542
[1/2] STEP 3/5: WORKDIR /go/src
--> Using cache 2cd201749a21dbdef077684622357f002f16e277671a1ea608ff9ee8580967a2
--> 2cd201749a2
[1/2] STEP 4/5: COPY . .
--> Using cache fc81df2e4c48496214159492c71b48236e2ec97ef8f777fb429a6a8d56cd9376
--> fc81df2e4c4
[1/2] STEP 5/5: RUN GOOS=linux CGO_ENABLED=0 go build -mod=vendor -o server .
--> Using cache 9326a789414fcdcfc3e03c26db0b49d7c02dc44c60a933e111e5c62745cd38f3
--> 9326a789414
[2/2] STEP 1/5: FROM registry.access.redhat.com/ubi8/ubi:latest
[2/2] STEP 2/5: WORKDIR /
--> Using cache f2b67a5c51104e9ab1eb4804251884913c0e34667a703e9dc8ed9bfa65f7d095
--> f2b67a5c511
[2/2] STEP 3/5: COPY --from=builder /go/src/server /usr/local/bin/server
--> Using cache 5a377698f410088344d248f10f249707fa7802f6317eb91c5dca43bd8daa07ea
--> 5a377698f41
[2/2] STEP 4/5: USER 65532:65532
--> Using cache 8ffdc894eb48342f08f74d7d84e3895bba7dd73e663fbc172d6775cd0bbe82b2
--> 8ffdc894eb4
[2/2] STEP 5/5: ENTRYPOINT ["/usr/local/bin/server"]
--> Using cache 1c66bc404a30459aae8fb37b52a7661ffe7c62925bea59ddaba4a3240608a5e6
[2/2] COMMIT amcdermo/ocpbugs-173-server:latest
--> 1c66bc404a3
Successfully tagged quay.io/amcdermo/ocpbugs-173-server:latest
Successfully tagged registry.int.frobware.com/aim/ocpbugs-173-server:latest
1c66bc404a30459aae8fb37b52a7661ffe7c62925bea59ddaba4a3240608a5e6
podman tag amcdermo/ocpbugs-173-server:latest quay.io/amcdermo/ocpbugs-173-server:latest
podman push quay.io/amcdermo/ocpbugs-173-server:latest
Getting image source signatures
Copying blob sha256:f284743ed5590bf870c2f0ca6f320f63dc28c0e7908cce7815339912ea386532
Copying blob sha256:5966005eac8d0b52bf676cd20f1ffb3435fe4d8245a3afadcd27b0b9e07c096b
Copying blob sha256:9936c6aaa811c2084fe2c1034e24cbecc3d3ac8db8fe987395723b19c678655b
Copying config sha256:1c66bc404a30459aae8fb37b52a7661ffe7c62925bea59ddaba4a3240608a5e6
Writing manifest to image destination
Storing signatures
make: Leaving directory '/home/aim/src/github.com/frobware/haproxy-hacks/OCPBUGS-173/server'
```

# Deploy websocket server and routes

```
$ oc process -f ./server/deployment.yaml | oc delete --ignore-not-found -f -
$ oc process -f ./server/deployment.yaml | oc apply -f -
```

Verify deployment:

```
$ oc logs ocpbugs-173-server-786f5d7cc7-gvzg9 -f
2022/08/26 13:32:31.696576 Listening securely on port 8443
2022/08/26 13:32:31.697159 Listening on port 8080
```

I have some Let's Encrypt certificates configured so I will use those
for the edge and reencrypt routes.

```
$ ./process-signed-routes.sh | oc delete --ignore-not-found -f -
$ ./process-signed-routes.sh | oc apply -f -

$ oc get routes
NAME                  HOST/PORT                                                  PATH   SERVICES             PORT   TERMINATION          WILDCARD
websocket-edge        websocket-edge-default.apps.ocp411.int.frobware.com        /      ocpbugs-173-server   8080   edge/Redirect        None
websocket-insecure    websocket-insecure-default.apps.ocp411.int.frobware.com    /      ocpbugs-173-server   8080                        None
websocket-reencrypt   websocket-reencrypt-default.apps.ocp411.int.frobware.com   /      ocpbugs-173-server   8443   reencrypt/Redirect   None
```

# Testing

We're going to use `wscat` for our testing, though see the later
section on using a browser for testing. If you don't have `wscat` use
the [create-wscat-container](create-wscat-container) script that will
create a fedora-based container with wscat installed.

## Testing with HTTP/2 disabled

We will use the default ingresscontroller for all our testing.

```
$ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=false
ingresscontroller.operator.openshift.io/default annotated
```

Wait for the new pods to roll out but, more importantly, wait for the
existing pods to exit `Terminating` state and no longer be listed.
Note: if you already had HTTP/2 disabled this will be a no-op.

```
$ oc get pods -n openshift-ingress
NAME                              READY   STATUS        RESTARTS   AGE
router-default-5d646c7b9b-4fpfc   2/2     Running       0          32s
router-default-5d646c7b9b-7xwmr   2/2     Running       0          67s
router-default-744f6c8bc4-8q7sd   1/2     Terminating   0          57m
router-default-744f6c8bc4-nd46p   1/2     Terminating   0          57m

$ oc get pods -n openshift-ingress
NAME                              READY   STATUS    RESTARTS   AGE
router-default-5d646c7b9b-4fpfc   2/2     Running   0          53s
router-default-5d646c7b9b-7xwmr   2/2     Running   0          88s
```

### Testing insecure route (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c ws://websocket-insecure-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.130.0.1:40018] Upgrade: [websocket]
< [10.130.0.1:40018] X-Forwarded-Host: [websocket-insecure-default.apps.ocp411.int.frobware.com]
< [10.130.0.1:40018] X-Forwarded-For: [192.168.7.203]
< [10.130.0.1:40018] Connection: [Upgrade]
< [10.130.0.1:40018] Sec-Websocket-Key: [EamO9e7XScCm3wVyfTz+mQ==]
< [10.130.0.1:40018] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< [10.130.0.1:40018] X-Forwarded-Port: [80]
< [10.130.0.1:40018] X-Forwarded-Proto: [http]
< [10.130.0.1:40018] Forwarded: [for=192.168.7.203;host=websocket-insecure-default.apps.ocp411.int.frobware.com;proto=http]
< [10.130.0.1:40018] Sec-Websocket-Version: [13]
< echo: headers
```

### Testing edge route (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c wss://websocket-edge-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.130.0.1:33954] Forwarded: [for=192.168.7.203;host=websocket-edge-default.apps.ocp411.int.frobware.com;proto=https]
< [10.130.0.1:33954] Sec-Websocket-Key: [HtYZtOOzH8Afb/oExbXUjw==]
< [10.130.0.1:33954] Connection: [Upgrade]
< [10.130.0.1:33954] Upgrade: [websocket]
< [10.130.0.1:33954] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< [10.130.0.1:33954] X-Forwarded-Proto: [https]
< [10.130.0.1:33954] Sec-Websocket-Version: [13]
< [10.130.0.1:33954] X-Forwarded-Host: [websocket-edge-default.apps.ocp411.int.frobware.com]
< [10.130.0.1:33954] X-Forwarded-Port: [443]
< [10.130.0.1:33954] X-Forwarded-For: [192.168.7.203]
< echo: headers
```

### Testing reencrypt route (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c wss://websocket-reencrypt-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.130.0.1:44220] Sec-Websocket-Version: [13]
< [10.130.0.1:44220] Upgrade: [websocket]
< [10.130.0.1:44220] X-Forwarded-Port: [443]
< [10.130.0.1:44220] X-Forwarded-Proto: [https]
< [10.130.0.1:44220] Forwarded: [for=192.168.7.203;host=websocket-reencrypt-default.apps.ocp411.int.frobware.com;proto=https]
< [10.130.0.1:44220] X-Forwarded-For: [192.168.7.203]
< [10.130.0.1:44220] Sec-Websocket-Key: [uIts876joiCI6qG+LQ7HbA==]
< [10.130.0.1:44220] Connection: [Upgrade]
< [10.130.0.1:44220] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< [10.130.0.1:44220] X-Forwarded-Host: [websocket-reencrypt-default.apps.ocp411.int.frobware.com]
< echo: headers
```

## Tesing with HTTP/2 enabled

```
$ oc -n openshift-ingress-operator annotate --overwrite ingresscontrollers/default ingress.operator.openshift.io/default-enable-http2=true
ingresscontroller.operator.openshift.io/default annotated
```

Wait for the new pods to roll out but, more importantly, wait for the
existing pods to exit `Terminating` state and no longer be listed:

```
$ oc get pods -n openshift-ingress
NAME                              READY   STATUS        RESTARTS   AGE
router-default-5d646c7b9b-4fpfc   1/2     Terminating   0          7m12s
router-default-5d646c7b9b-7xwmr   2/2     Terminating   0          7m47s
router-default-6c8cc7f9dc-jskjw   2/2     Running       0          42s
router-default-6c8cc7f9dc-vt4ng   2/2     Running       0          7s

$ oc get pods -n openshift-ingress
NAME                              READY   STATUS    RESTARTS   AGE
router-default-6c8cc7f9dc-jskjw   2/2     Running   0          84s
router-default-6c8cc7f9dc-vt4ng   2/2     Running   0          49s
```

### Testing insecure route (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c ws://websocket-insecure-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.131.0.1:51414] X-Forwarded-For: [192.168.7.203]
< [10.131.0.1:51414] Connection: [Upgrade]
< [10.131.0.1:51414] X-Forwarded-Host: [websocket-insecure-default.apps.ocp411.int.frobware.com]
< [10.131.0.1:51414] X-Forwarded-Port: [80]
< [10.131.0.1:51414] X-Forwarded-Proto: [http]
< [10.131.0.1:51414] Forwarded: [for=192.168.7.203;host=websocket-insecure-default.apps.ocp411.int.frobware.com;proto=http]
< [10.131.0.1:51414] Sec-Websocket-Version: [13]
< [10.131.0.1:51414] Sec-Websocket-Key: [juk5CzEPGr13RLGUPVEIMw==]
< [10.131.0.1:51414] Upgrade: [websocket]
< [10.131.0.1:51414] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< echo: headers
```

### Testing edge route (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c wss://websocket-edge-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.131.0.1:53316] X-Forwarded-Port: [443]
< [10.131.0.1:53316] Forwarded: [for=192.168.7.203;host=websocket-edge-default.apps.ocp411.int.frobware.com;proto=https]
< [10.131.0.1:53316] X-Forwarded-For: [192.168.7.203]
< [10.131.0.1:53316] Sec-Websocket-Version: [13]
< [10.131.0.1:53316] Sec-Websocket-Key: [7QQ7yaDWIUq7FLI6Dxc15w==]
< [10.131.0.1:53316] X-Forwarded-Host: [websocket-edge-default.apps.ocp411.int.frobware.com]
< [10.131.0.1:53316] X-Forwarded-Proto: [https]
< [10.131.0.1:53316] Connection: [Upgrade]
< [10.131.0.1:53316] Upgrade: [websocket]
< [10.131.0.1:53316] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< echo: headers
```

### Testing reencrypt route (expect **FAILURE**)

```
$ ./wscat --no-color -n -L -c wss://websocket-reencrypt-default.apps.ocp411.int.frobware.com/echo
error: Unexpected server response: 400
```

Looking at the pods logs we see:

```
$ oc logs ocpbugs-173-server-786f5d7cc7-gvzg9
2022/08/26 13:53:43.902145 [10.131.0.1:48488] Host: [websocket-reencrypt-default.apps.ocp411.int.frobware.com]
2022/08/26 13:53:43.902214 [10.131.0.1:48488] Sec-Websocket-Version: [13]
2022/08/26 13:53:43.902226 [10.131.0.1:48488] Sec-Websocket-Key: [Whty9xbKN4t0B+Dxl05c3A==]
2022/08/26 13:53:43.902237 [10.131.0.1:48488] X-Forwarded-Port: [443]
2022/08/26 13:53:43.902247 [10.131.0.1:48488] X-Forwarded-Proto: [https]
2022/08/26 13:53:43.902257 [10.131.0.1:48488] Forwarded: [for=192.168.7.203;host=websocket-reencrypt-default.apps.ocp411.int.frobware.com;proto=https]
2022/08/26 13:53:43.902267 [10.131.0.1:48488] X-Forwarded-For: [192.168.7.203]
2022/08/26 13:53:43.902277 [10.131.0.1:48488] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
2022/08/26 13:53:43.902287 [10.131.0.1:48488] X-Forwarded-Host: [websocket-reencrypt-default.apps.ocp411.int.frobware.com]
2022/08/26 13:53:43.902303 upgrade:websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header
```

Using websockets and a reencrypt route (i.e., H2->H2) is not supported
in haproxy 2.2. This is the upstream issue
https://github.com/haproxy/haproxy/issues/162. If you look at the
error status we get back (i.e., 400) then this matches the description
in issue 162. However, Websockets over HTTP/2 is now a feature of
[HAProxy 2.4](https://www.haproxy.com/blog/announcing-haproxy-2-4/).

## Testing with HTTP/2 enabled and HAProxy 2.4

Let's validate 2.4.

We need to disable both the CVO and the cluster-ingress-operator as we
will be replacing the router image.

```
$ oc scale --replicas 0 -n openshift-cluster-version deployments/cluster-version-operator
$ oc scale --replicas 0 -n openshift-ingress-operator deployments ingress-operator
$ oc -n openshift-ingress set image deployment/router-default router=quay.io/amcdermo/openshift-router-haproxy-2.4.18:latest
```

New pods will rollout:

```
$ oc get pods -n openshift-ingress -w
NAME                              READY   STATUS              RESTARTS   AGE
router-default-6c8cc7f9dc-jskjw   2/2     Terminating         0          14m
router-default-6c8cc7f9dc-vt4ng   2/2     Running             0          13m
router-default-744f6c8bc4-mmh8k   0/2     ContainerCreating   0          2s
router-default-744f6c8bc4-mmh8k   1/2     Running             0          6s
router-default-744f6c8bc4-mmh8k   1/2     Running             0          6s
router-default-744f6c8bc4-mmh8k   2/2     Running             0          7s
```

As before wait for just two router pods:

```
$ oc get pods -n openshift-ingress
NAME                              READY   STATUS              RESTARTS   AGE
router-default-6c8cc7f9dc-jskjw   1/2     Terminating         0          15m
router-default-6c8cc7f9dc-vt4ng   2/2     Terminating         0          14m
router-default-744f6c8bc4-mmh8k   2/2     Running             0          51s
router-default-744f6c8bc4-pjt6d   0/2     ContainerCreating   0          14s

$ oc get pods -n openshift-ingress
NAME                              READY   STATUS    RESTARTS   AGE
router-default-744f6c8bc4-mmh8k   2/2     Running   0          2m11s
router-default-744f6c8bc4-pjt6d   2/2     Running   0          94s
```

Now let's rerun the reencrypt test Note: the only change we have made
is to swap out haproxy 2.2 for 2.4 but let's verify that:

```
$ oc rsh -n openshift-ingress router-default-744f6c8bc4-mmh8k /usr/sbin/haproxy -v
Defaulted container "router" out of: router, logs
HAProxy version 2.4.18-1d80f18 2022/07/27 - https://haproxy.org/
Status: long-term supported branch - will stop receiving fixes around Q2 2026.
Known bugs: http://www.haproxy.org/bugs/bugs-2.4.18.html
Running on: Linux 4.18.0-372.19.1.el8_6.x86_64 #1 SMP Mon Jul 18 11:14:02 EDT 2022 x86_64
```

### Testing reencrypt route with HAProxy 2.4 (expect **SUCCESS**)

```
$ ./wscat --no-color -n -L -c wss://websocket-reencrypt-default.apps.ocp411.int.frobware.com/echo
Connected (press CTRL+C to quit)
> foo
< echo: foo
> bar
< echo: bar
> headers
< [10.130.0.1:32886] Upgrade: [websocket]
< [10.130.0.1:32886] X-Forwarded-Host: [websocket-reencrypt-default.apps.ocp411.int.frobware.com]
< [10.130.0.1:32886] Sec-Websocket-Version: [13]
< [10.130.0.1:32886] Sec-Websocket-Key: [1hab4pAgAqOJEBDIHu2ScA==]
< [10.130.0.1:32886] Connection: [Upgrade]
< [10.130.0.1:32886] Sec-Websocket-Extensions: [permessage-deflate; client_max_window_bits]
< [10.130.0.1:32886] X-Forwarded-Port: [443]
< [10.130.0.1:32886] X-Forwarded-Proto: [https]
< [10.130.0.1:32886] Forwarded: [for=192.168.7.203;host=websocket-reencrypt-default.apps.ocp411.int.frobware.com;proto=https]
< [10.130.0.1:32886] X-Forwarded-For: [192.168.7.203]
< echo: headers
```

# Testing via a web browser

If you don't want to use `wscat` you can test via a browser.

Vist one of the (equivalent) URLs:

- http://websocket-insecure-default.apps.ocp411.int.frobware.com
- https://websocket-edge-default.apps.ocp411.int.frobware.com
- https://websocket-reencrypt-default.apps.ocp411.int.frobware.com

The web page allows you to "Open" a connection to the websocket
server. Once sucessfully opened, you can send "commands" which will
get echoed back to the web page. If you send the command "headers" it
will reply with all the headers associated with the request as
received by the backend websocket server.

![Web Browser Testing](screenshots/browser.png?raw=true "Web Browser Testing")
