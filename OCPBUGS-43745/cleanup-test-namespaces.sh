#!/usr/bin/env bash

oc delete namespaces $(oc get namespaces --no-headers -A | grep route-service-switch | awk '{print $1}')

