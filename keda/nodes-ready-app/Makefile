TAG ?= amcdermo/nodes-ready-app

.PHONY: build
build:
	CGO_ENABLED=0 go build -mod=vendor -o nodes-ready-app main.go

image: fmt vet vendor build
	docker build -f Containerfile -t $(TAG)

vendor:
	go mod vendor
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...
