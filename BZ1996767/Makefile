BRANCH	  := $(shell git rev-parse --abbrev-ref HEAD)
HASH	  := $(shell git rev-parse HEAD | colrm 8)
TAG	  := $(BRANCH)-g$(HASH)
IMAGENAME := amcdermo/bz1996767-server

build:
	buildah bud -t amcdermo/$(IMAGENAME):$(TAG) -f Dockerfile .

push:
	podman tag amcdermo/$(IMAGENAME):$(TAG) quay.io/$(IMAGENAME):$(TAG)
	podman push quay.io/$(IMAGENAME):$(TAG)
