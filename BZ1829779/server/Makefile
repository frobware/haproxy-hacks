build:
	go build -o server server.go
	docker build . -t frobware/bz1829779-server

push: build
	docker push frobware/bz1829779-server
