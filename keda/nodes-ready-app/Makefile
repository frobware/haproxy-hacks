TAG ?= amcdermo/nodes-ready-app

.PHONY: build
build:
	CGO_ENABLED=0 go build -mod=vendor -o nodes-ready-app .

image: $(HOME)/.config/containers/policy.json fmt vet vendor build
	$(RM) nodes-ready-app
	podman build -f Containerfile -t $(TAG)

push-image:
	podman tag $(TAG) quay.io/$(TAG)
	podman push quay.io/$(TAG)

vendor:
	go mod vendor
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

$(HOME)/.config/containers/policy.json: policy.json | $(HOME)/.config/containers
	cp policy.json $@

$(HOME)/.config/containers:
	mkdir -p $@
