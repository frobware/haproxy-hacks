# USE_DEV_INIT = --init

run:
	go build -o reaper main.go && \
		docker build -t reaper-test . && \
		docker run $(USE_DEV_INIT) -it --rm reaper-test
