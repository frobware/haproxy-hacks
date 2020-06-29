#!/bin/bash

for i in $(seq 1 ${1:-10}); do
    oc delete -n lots-of-routes route helloworld-${i}
done

for i in $(seq 1 ${1:-10}); do
    oc expose -n lots-of-routes service helloworld-${i}
done
