REGISTRY ?= registry.int.frobware.com
IMAGE_ORG ?= aim
IMAGE_NAME ?= perf-server
IMAGE_TAG ?= latest
IMAGE_REF ?= $(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_TAG)

build:
	CGO_ENABLED=0 go build -mod=vendor -o server .

image: fmt vet vendor build
	$(RM) server
	podman build -f Containerfile -t $(IMAGE_REF)

push-image: image
	podman tag $(IMAGE_REF) $(REGISTRY)/$(IMAGE_REF)
	podman push $(REGISTRY)/$(IMAGE_REF)

vendor:
	go mod vendor
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: build push-image vendor fmt vet
