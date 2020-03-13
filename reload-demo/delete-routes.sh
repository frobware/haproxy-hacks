#!/bin/bash

for i in $(seq 1 ${1:-10}); do
    oc delete route helloworld-${i}
done
