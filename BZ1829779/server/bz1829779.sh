#!/usr/bin/env bash

go_mod="$(sed -e 's/^/        /' go.mod    | sed '/^[[:space:]]*$/d')"
go_sum="$(sed -e 's/^/        /' go.sum    | sed '/^[[:space:]]*$/d')"

go_server="$(sed -e 's/^/        /' server.go | sed '/^[[:space:]]*$/d')"

cat <<EOF
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: bz1829779
  spec:
    selector:
      app: bz1829779
    ports:
      - port: 3264
        name: http
        targetPort: 3264
        protocol: TCP
- apiVersion: v1
  kind: ConfigMap
  labels:
    app: bz1829779
  metadata:
    name: bz1829779-src-config
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
    name: bz1829779
    labels:
      app: bz1829779
  spec:
    replicas: ${REPLICAS:-2}
    template:
      metadata:
        name: bz1829779
        labels:
          app: bz1829779
      spec:
        containers:
        - image: golang:1.14
          name: server
          command: ["go", "run", "/go/src/server.go"]
          env:
          - name: BUSY_TIMEOUT
            value: "${BUSY_TIMEOUT:-0}"
          - name: GO111MODULE
            value: "auto"
          - name: GOCACHE
            value: "/tmp"
          volumeMounts:
          - name: bz1829779-src-volume
            mountPath: /go/src
          readinessProbe:
            httpGet:
              path: /healthz
              port: 3264
            initialDelaySeconds: 3
            periodSeconds: 3
        volumes:
        - name: bz1829779-src-volume
          configMap:
            name: bz1829779-src-config
    selector:
      matchLabels:
        app: bz1829779
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: bz1829779
    name: bz1829779-edge
  spec:
    port:
      targetPort: 3264
    tls:
      termination: edge
      insecureEdgeTerminationPolicy: Redirect
      key: |-
      certificate: |-
    to:
      kind: Service
      name: bz1829779
      weight: 100
    wildcardPolicy: None
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: bz1829779
    name: bz1829779-insecure
  spec:
    port:
      targetPort: 3264
    to:
      kind: Service
      name: bz1829779
      weight: 100
    wildcardPolicy: None
EOF
