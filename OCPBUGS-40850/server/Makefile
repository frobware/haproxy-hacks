image: server
	podman build -t ocpbugs40850-server .
	podman tag localhost/ocpbugs40850-server:latest quay.io/amcdermo/ocpbugs40850-server
	podman push quay.io/amcdermo/ocpbugs40850-server

server: server.go
	CGO_ENABLED=0 go build -o server server.go

regress:
	./regress.sh 1051 current
	diff -r tests/current tests/baseline
