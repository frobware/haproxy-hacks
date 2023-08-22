# If this Makefile changes, then everthing should be rebuilt.
.EXTRA_PREREQS := $(abspath $(lastword $(MAKEFILE_LIST)))

REGISTRY ?= quay.io
IMAGE ?= amcdermo/openshift-router-ocpbugs16707-server

server: server.go
	CGO_ENABLED=0 go build server.go
	podman build -t localhost/$(IMAGE) .
	podman tag localhost/$(IMAGE):latest $(REGISTRY)/$(IMAGE)
	podman push $(REGISTRY)/$(IMAGE)

clean:
	$(RM) -f server
