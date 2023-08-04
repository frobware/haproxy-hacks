server: server.go
	CGO_ENABLED=0 go build server.go
	podman build -t ocpbugs13966-server .
	podman tag localhost/ocpbugs13966-server:latest registry.int.frobware.com/ocpbugs13966/ocpbugs13966-server
	podman push registry.int.frobware.com/ocpbugs13966/ocpbugs13966-server

clean:
	$(RM) -f server
