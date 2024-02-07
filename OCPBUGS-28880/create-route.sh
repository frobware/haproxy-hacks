#!/usr/bin/env bash

set -e

oc create route edge --service normal1

oc annotate route normal1 haproxy.router.openshift.io/balance=roundrobin
oc annotate route normal1 haproxy.router.openshift.io/disable_cookies=true
