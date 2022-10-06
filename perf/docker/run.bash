#!/usr/bin/env bash

docker-compose up -t 1 --scale nginx=${1:-100}
