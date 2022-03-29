BRANCH	  := $(shell git rev-parse --abbrev-ref HEAD)
HASH	  := $(shell git rev-parse HEAD | colrm 8)
TAG	  := $(BRANCH)-g$(HASH)
IMAGENAME := amcdermo/bz2068484-server

build:
	buildah bud -t amcdermo/$(IMAGENAME):$(TAG) -f Dockerfile .

push:
	podman tag amcdermo/$(IMAGENAME):$(TAG) quay.io/$(IMAGENAME):$(TAG)
	podman push quay.io/$(IMAGENAME):$(TAG)

server:
	GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server
