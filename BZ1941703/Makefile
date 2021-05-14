all:
	docker build -t frobware/cio .
	docker tag frobware/cio:latest quay.io/frobware/cio
	docker push quay.io/frobware/cio:latest

rollover:
	-oc delete -f ./server/server.yaml
	oc create -f ./server/server.yaml

refresh: all rollover
