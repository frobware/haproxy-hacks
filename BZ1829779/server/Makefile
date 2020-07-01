build: server
	docker build . -t frobware/bz1829779-server

server: server.go
	go build -o server server.go

push: build
	docker push frobware/bz1829779-server
