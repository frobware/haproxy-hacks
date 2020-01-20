DOCKER ?= docker

run: build
	$(DOCKER) run -it --rm --init --name prometheus-test -p 8089:8089 -p 8404:8404 prometheus-test

build:
	$(DOCKER) build -t prometheus-test .

reload:
	$(DOCKER) exec -
