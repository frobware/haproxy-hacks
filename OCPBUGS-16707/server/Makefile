# If this Makefile changes, then everthing should be rebuilt.
.EXTRA_PREREQS := $(abspath $(lastword $(MAKEFILE_LIST)))

server: server.go
	CGO_ENABLED=0 go build server.go
	podman build -t server .
	podman tag localhost/server:latest registry.int.frobware.com/ocpbugs16707/server
	podman push registry.int.frobware.com/ocpbugs16707/server

clean:
	$(RM) -f server
