#!/usr/bin/env bash

go_mod="$(sed -e 's/^/        /' go.mod    | sed '/^[[:space:]]*$/d')"
go_sum="$(sed -e 's/^/        /' go.sum    | sed '/^[[:space:]]*$/d')"

go_server="$(sed -e 's/^/        /' server/server.go | sed '/^[[:space:]]*$/d')"

cat <<EOF
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: bz1841454
  spec:
    selector:
      app: bz1841454
    ports:
      - port: 8080
        name: http
        targetPort: 8080
        protocol: TCP
- apiVersion: v1
  kind: ConfigMap
  labels:
    app: bz1841454
  metadata:
    name: src-config
  data:
    go.mod: |
$go_mod
    go.sum: |
$go_sum
    server.go: |
$go_server
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bz1841454
    labels:
      app: bz1841454
  spec:
    replicas: ${REPLICAS:-2}
    template:
      metadata:
        name: bz1841454
        labels:
          app: bz1841454
      spec:
        initContainers:
        - image: golang:1.14
          name: builder
          command: ["/bin/bash", "-c"]
          args:
          - set -e;
            cd /go/src;
            go build -o /workdir/server -x -v -mod=readonly server.go;
          env:
          - name: GO111MODULE
            value: "auto"
          - name: GOCACHE
            value: "/tmp"
          volumeMounts:
          - name: src-volume
            mountPath: /go/src
          - name: workdir
            mountPath: /workdir
        volumes:
        - name: src-volume
          configMap:
            name: src-config
        - name: workdir
          emptyDir: {}
        containers:
        - image: golang:1.14
          name: server
          command: ["/workdir/server"]
          env:
          - name: BUSY_TIMEOUT
            value: "${BUSY_TIMEOUT:-0}"
          volumeMounts:
          - name: workdir
            mountPath: /workdir
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 3
        volumes:
        - name: src-volume
          configMap:
            name: src-config
        - name: workdir
          emptyDir: {}
    selector:
      matchLabels:
        app: bz1841454
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: bz1841454
    name: bz1841454-edge
  spec:
    port:
      targetPort: 8080
    tls:
      termination: edge
      insecureEdgeTerminationPolicy: Redirect
    to:
      kind: Service
      name: bz1841454
      weight: 100
    wildcardPolicy: None
EOF
