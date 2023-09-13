# https://issues.redhat.com/browse/OCPBUGS-18820

## Repro steps

    oc new-app --name=http2-echo --image=quay.io/rhn_support_nsu/kalmhq-echoserver
	oc expose deployment http2-echo --port=8002
	oc expose svc http2-echo 

## Check whether http2 is enabled/disabled

	oc get ingresscontroller -n openshift-ingress-operator default -o yaml | grep -i http2
	oc get ingresses.config/cluster -o yaml | grep -i http2

## Test

	curl -v --http1.1 -I http2-echo-ocpbugs18820.apps.ocp413.int.frobware.com
