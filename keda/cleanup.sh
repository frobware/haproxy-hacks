#!/usr/bin/env bash

oc delete jobs --field-selector status.successful=1 
oc delete -f triggerauthentication.yaml
oc delete -f scaledobject.yaml
oc delete -f deployment.yaml
oc delete -f secret.yaml
oc delete -f rolebinding.yaml
oc delete -f role.yaml
