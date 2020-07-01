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
        -----BEGIN EC PRIVATE KEY-----
        MHcCAQEEIAW+ecg2cZR47ItbI898N3nJduh9UJNv+b0cOwH/Z1BEoAoGCCqGSM49
        AwEHoUQDQgAEx0/5sEgiUPFdcbd4dSllkul8s68RQ5WxIjfwWYMdfYLiLLqP1lkz
        4UYpwAW/t63qBx3jRhPgkUxh5saJP9Qu5Q==
        -----END EC PRIVATE KEY-----
      certificate: |-
        -----BEGIN CERTIFICATE-----
        MIIBgTCCASagAwIBAgIRALutWdExjxX8fWljW+lcYbswCgYIKoZIzj0EAwIwJDEQ
        MA4GA1UEChMHUmVkIEhhdDEQMA4GA1UEAxMHUm9vdCBDQTAgFw0yMDA1MTExMDU2
        NThaGA8yMTIwMDQxNzEwNTY1OFowJjEQMA4GA1UEChMHUmVkIEhhdDESMBAGA1UE
        AwwJdGVzdF9jZXJ0MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEx0/5sEgiUPFd
        cbd4dSllkul8s68RQ5WxIjfwWYMdfYLiLLqP1lkz4UYpwAW/t63qBx3jRhPgkUxh
        5saJP9Qu5aM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMB
        MAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDSQAwRgIhAOIx8885y8tX/Vv94UGx
        hWC/O1Hzi15kOT0WQ/UKUMjMAiEA40uW9P6k+i1cDwgfBBMzgDFQa9GAb4FqM8Wr
        PaUMdqg=
        -----END CERTIFICATE-----
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
