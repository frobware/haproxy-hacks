# USE_DEV_INIT = --init

DOCKER ?= podman

run:
	go build -o reaper main.go && \
		$(DOCKER) build -t reaper-test . && \
		$(DOCKER) run $(USE_DEV_INIT) -it --rm reaper-test
