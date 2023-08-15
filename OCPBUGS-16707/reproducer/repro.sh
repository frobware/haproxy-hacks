#!/usr/bin/env bash

oc delete --ignore-not-found -n ocpbugs16707 route route1 route2 route3

oc create -f ./reproducer/route1.yaml
sleep 1
oc create -f ./reproducer/route2.yaml
sleep 1
oc create -f ./reproducer/route3.yaml
sleep 1
oc get routes route1 route2 route3

oc describe route route3 -n ocpbugs16707
