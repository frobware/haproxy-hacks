TAG ?= amcdermo/ocpbugs-173-server

build:
	CGO_ENABLED=0 go build -mod=vendor -o server .

image: fmt vet vendor build
	$(RM) server
	podman build -f Containerfile -t $(TAG)

push-image: image
	podman tag $(TAG) quay.io/$(TAG)
	podman push quay.io/$(TAG)

vendor:
	go mod vendor
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: build push-image vendor fmt vet
