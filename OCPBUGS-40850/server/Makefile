GOFILES := $(wildcard *.go)
NON_TEST_GOFILES := $(filter-out %_test.go, $(GOFILES))

HTTP_PORT ?= 1051
HTTPS_PORT ?= 1052

export HTTP_PORT
export HTTPS_PORT

server: server.go
	CGO_ENABLED=0 go build -o server $(NON_TEST_GOFILES)

image: server
	podman build -t ocpbugs40850-server .
	podman tag localhost/ocpbugs40850-server:latest quay.io/amcdermo/ocpbugs40850-server
	podman push quay.io/amcdermo/ocpbugs40850-server

.PHONY: regress
regress:
	$(RM) -r tests/current
	mkdir -p tests/current
	./regress.sh 1051 tests/current
	diff tests/current tests/baseline

run-server: server
	$(SUDO) ./server
