test.yaml: check-server-build
	./test.sh > $@

check-server-build:
	go build -o /dev/null ./server

.PHONY: \
	test.yaml \
	check-server-build \
