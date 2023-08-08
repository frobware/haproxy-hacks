# If this Makefile changes, then everthing should be rebuilt.
.EXTRA_PREREQS := $(abspath $(lastword $(MAKEFILE_LIST)))

all: server

NAMESPACE := $(shell oc project -q)
SUBDOMAIN := $(shell oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

server: server.go index.html
	CGO_ENABLED=0 go build server.go
	podman build -t server .
	podman tag localhost/server:latest registry.int.frobware.com/ocpbugs12858/server
	podman push registry.int.frobware.com/ocpbugs12858/server

index.html: index.template
	sed -e 's/SUBDOMAIN/$(SUBDOMAIN)/g' -e 's/NAMESPACE/$(NAMESPACE)/g' $< > $@

verify:
	echo NAMESPACE=$(NAMESPACE)
	echo SUBDOMAIN=$(SUBDOMAIN)

clean:
	$(RM) -f index.html server
