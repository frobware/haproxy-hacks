#!/usr/bin/env bash

while :
do
    oc delete route helloworld-0-edge
    sleep 1
    oc expose service helloworld-0 --name=helloworld-1-edge --port 8080
done
