all: server

NAMESPACE := $(shell oc project -q)
SUBDOMAIN := $(shell oc get -n openshift-ingress-operator ingresscontroller/default -o yaml | yq .status.domain)

server: server.go index.html
	CGO_ENABLED=0 go build server.go
	podman build -t server .
	podman tag localhost/server:latest registry.int.frobware.com/destca/server
	podman push registry.int.frobware.com/destca/server

index.html: index.m4
	m4 -DSUBDOMAIN=$(SUBDOMAIN) -DNAMESPACE=$(NAMESPACE) $< > $@

verify:
	echo NAMESPACE=$(NAMESPACE)
	echo SUBDOMAIN=$(SUBDOMAIN)

clean:
	$(RM) -f index.html server
