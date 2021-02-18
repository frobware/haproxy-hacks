# Watch events

```shell
oc get events --no-headers --sort-by=.metadata.creationTimestamp --watch
```

# Watch for pod restarts

```shell
oc get pods -o wide -w
```


# Set proxy

```
oc set env deployment/bz1929235 HTTP_PROXY=http://<PROXY>:3128
oc set env deployment/bz1929235 HTTPS_PROXY=http://<PROXY>:3128
oc set env deployment/bz1929235 NO_PROXY=...
```
