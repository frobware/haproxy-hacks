#!/usr/bin/env bash

while :
do
    oc delete route helloworld-8-insecure
    sleep 5
    oc expose service helloworld-1 --name=helloworld-8-insecure --port 8080
done
