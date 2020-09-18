#!/bin/bash

for i in $(seq 1 ${1:-10}); do
    oc ${OP:-expose} service helloworld-${i}
done
