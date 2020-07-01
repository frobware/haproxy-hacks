bz1829779-hostnetwork.yaml: server
	./server-hostnetwork.sh > $@

server: server.go
	go build -o server server.go

check-server-build:
	go build -o server server.go

build: server
	docker build . -t frobware/bz1829779-server

push: build
	docker push frobware/bz1829779-server

.PHONY: bz1829779-hostnetwork.yaml
