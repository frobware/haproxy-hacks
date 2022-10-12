#!/usr/bin/env bash

docker-compose up --remove-orphans -t 1 --scale nginx=${1:-100}
