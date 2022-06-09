- Useful links

- https://github.com/zroubalik/keda-openshift-examples/tree/main/prometheus
- https://access.redhat.com/articles/6718611
- https://issues.redhat.com/browse/RHSTOR-1938

* Install steps

    oc create -f ./cluster-role.yaml
    oc adm policy add-role-to-user thanos-metrics-reader -z SERVICE_ACCOUNT --role-namespace=openshift-ingress-operator

    SECRET=$(oc get secret -n openshift-storage | grep  <SERVICE_ACCOUNT>-token | head -n 1 | awk '{print $1 }')




