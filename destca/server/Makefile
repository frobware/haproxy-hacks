all: server

server: server.go index.html
	CGO_ENABLED=0 go build server.go
	podman build -t server .
	podman tag localhost/server:latest registry.int.frobware.com/destca/server
	podman push registry.int.frobware.com/destca/server

index.html: index.m4
	m4 -DSUBDOMAIN=ocp414.int.frobware.com $< > $@
