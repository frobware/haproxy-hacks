test.yaml: check-server-build
	./test.sh > $@

check-server-build:
	go build -o server ./server/server.go

.PHONY: \
	test.yaml \
	check-server-build \
