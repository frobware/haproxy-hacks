#!/bin/bash

# for i in $(seq 1 ${1:-10}); do
#     oc delete -n lots-of-routes route helloworld-${i}
# done

# for i in $(seq 1 ${1:-10}); do
#     oc expose -n lots-of-routes service helloworld-${i}
# done

while :
do
    oc delete service/my-service
    sleep 1.5
    oc expose pod helloworld-1 --type=ClusterIP --name=my-service --port 8080
    sleep 2.5
done
