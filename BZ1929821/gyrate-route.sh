#!/usr/bin/env bash

# for i in $(seq 1 ${1:-10}); do
#     oc delete -n lots-of-routes route helloworld-${i}
# done

# for i in $(seq 1 ${1:-10}); do
#     oc expose -n lots-of-routes service helloworld-${i}
# done

while :
do
    oc delete route helloworld-1-insecure
    sleep 1.5
    oc expose service helloworld-1 --name=helloworld-1-insecure --port 8080
    sleep 2.5
done
