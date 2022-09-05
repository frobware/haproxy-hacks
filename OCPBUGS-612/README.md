# DNS test

https://issues.redhat.com/browse/OCPBUGS-612

# Setup

    oc process -f deployment-dnstest.yaml | oc create -f -

This will create a pod with 5 containers.

Each each container asserts that it can resolve the address:

    "_grpc._tcp.prometheus-operated.openshift-monitoring.svc.cluster.local"

## Containers

| container                    | purpose                                           |
|------------------------------|---------------------------------------------------|
| nslookup                     | uses nslookup(1) to resolve addresses             |
| thanos-miekgdns              | uses miekgdns thanos resolver                     |
| thanos-golang                | uses golang thanos resolver                       |
| golang-nslookup-cgo-enabled  | uses net.LookupHost() compiled with CGO_ENABLED=1 |
| golang-nslookup-cgo-disabled | uses net.LookupHost() compiled with CGO_ENABLED=0 |


