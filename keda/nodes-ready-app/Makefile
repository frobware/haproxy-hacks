TAG ?= frobware/nodes-ready-app

.PHONY: build

build:
	go build -mod=vendor -o nodes-ready-app main.go

image: build
	docker build -f Containerfile -t $(TAG)
