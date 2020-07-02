all: bz1829779-hostnetwork.yaml bz1829779.yaml

bz1829779-hostnetwork.yaml: bz1829779-hostnetwork.sh | server
	./bz1829779-hostnetwork.sh > $@

bz1829779.yaml: bz1829779.sh | server
	./bz1829779.sh > $@

server: server.go
	go build -o server server.go

build: server
	docker build . -t frobware/bz1829779-server

push: build
	docker push frobware/bz1829779-server

clean:
	$(RM) *.yaml
