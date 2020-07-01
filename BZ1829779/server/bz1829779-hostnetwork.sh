#!/usr/bin/env bash

go_mod="$(sed -e 's/^/        /' go.mod    | sed '/^[[:space:]]*$/d')"
go_sum="$(sed -e 's/^/        /' go.sum    | sed '/^[[:space:]]*$/d')"
go_server="$(sed -e 's/^/        /' server.go | sed '/^[[:space:]]*$/d')"

cat <<EOF
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: ConfigMap
  labels:
    app: bz1829779-hostnetwork
  metadata:
    name: bz1829779-hostnetwork-src-config
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
    labels:
      app: bz1829779-hostnetwork
    name: bz1829779-hostnetwork
  spec:
    replicas: ${REPLICAS:-2}
    template:
      metadata:
        labels:
          app: bz1829779-hostnetwork
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
          - name: bz1829779-hostnetwork-src-volume
            mountPath: /go/src
          readinessProbe:
            httpGet:
              path: /healthz
              port: 3264
            initialDelaySeconds: 3
            periodSeconds: 3
        volumes:
        - name: bz1829779-hostnetwork-src-volume
          configMap:
            name: bz1829779-hostnetwork-src-config
        hostNetwork: true
        nodeSelector:
          role: node
          router: enabled
        securityContext: {}
        restartPolicy: Always
        serviceAccount: router
        serviceAccountName: router
        clusterIP: None
        dnsPolicy: ClusterFirstWithHostNet
    selector:
      matchLabels:
        app: bz1829779-hostnetwork
EOF
