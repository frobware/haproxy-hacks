#!/usr/bin/env bash

curl -k -L -I https://"$(oc get routes bz2044682-edge -o json | jq -r '.spec.host')"
