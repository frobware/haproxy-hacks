#!/usr/bin/env bash

oc delete --ignore-not-found -n ocpbugs16707 route route1 route2 route3

for i in 1 2 3; do
    oc create -f ./reproducer/route${i}.yaml
    sleep 2
done

oc describe route route3 -n ocpbugs16707
