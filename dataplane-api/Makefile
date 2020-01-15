DOCKER ?= docker

run:
	$(DOCKER) build -t dataplane-api-test . && \
	$(DOCKER) run -p 5555:5555 dataplane-api-test
