NAME = $(shell basename $$PWD)

.PHONY: build run clean

build:
	docker build --rm=false --tag frobware/$(NAME) .

run:
	docker run -v /tmp:/target frobware/$(NAME)

clean:
	-docker rmi frobware/$(NAME)
