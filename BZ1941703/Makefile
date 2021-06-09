HOST = $(shell oc get routes -o json | jq -r '.items[0].spec.host')

all:
	podman build -t amcdermo/bz1941703 .
	podman tag amcdermo/bz1941703:latest quay.io/amcdermo/bz1941703
	podman push quay.io/amcdermo/bz1941703:latest

rollover:
	-oc delete -f ./server/server.yaml
	oc create -f ./server/server.yaml

refresh: all rollover

run-client:
	go run client-reproducer/main.go -addr $(HOST):80

verify:
	@echo HOST=$(HOST)
