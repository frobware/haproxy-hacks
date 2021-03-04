#!/bin/bash

for i in $(seq 1 ${1:-10}); do
cat <<-EOF
---
apiVersion: v1
kind: List
items:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-1
    name: helloworld-${i}-edge
  spec:
    port:
      targetPort: 8080
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
      name: helloworld-1
      weight: 100
    wildcardPolicy: None
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    labels:
      app: helloworld-1
    name: helloworld-${i}-insecure
  spec:
    port:
      targetPort: 8080
    to:
      kind: Service
      name: helloworld-1
      weight: 100
    wildcardPolicy: None
EOF
done
