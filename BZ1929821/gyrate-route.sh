#!/usr/bin/env bash

while :
do
    oc delete route helloworld-1-insecure
    sleep 1.5
    oc expose service helloworld-1 --name=helloworld-1-insecure --port 8080
    sleep 2.5
done
