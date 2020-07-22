#!/usr/bin/env bash

./generate-pods.sh ${1:-100} | oc ${OP:-apply} -f -
./generate-services.sh ${1:-100} | oc ${OP:-apply} -f -
./expose-services.sh ${1:-100}
