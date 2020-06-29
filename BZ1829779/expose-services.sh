#!/bin/bash

for i in $(seq 1500 ${1:-10}); do
    oc expose service helloworld-${i}
done
