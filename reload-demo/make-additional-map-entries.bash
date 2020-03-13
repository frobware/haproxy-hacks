#!/usr/bin/bash

for i in $(seq 3001 13000); do
    cat <<EOF
^helloworld-${i}-default\.apps\.amcdermo-202003-13-0951\.devcluster\.openshift\.com(:[0-9]+)?(/.*)?$ be_http:default:helloworld-${i}
EOF
done
