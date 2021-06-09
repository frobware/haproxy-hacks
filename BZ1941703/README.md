# Usage

## Deploy workload

    oc new-project lot-of-routes
	./generate-pods-and-services.sh | oc apply -f -
	./generate-routes.sh 0 600 | oc apply -f -

## Deploy websocket backend

	oc new-project ws
    oc create -f ./server/server.yaml

## Force reload

	./gyrate-route.sh

## Run websocket client

	make run-client
