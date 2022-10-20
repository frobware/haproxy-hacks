#!/usr/bin/env bash

set -eu

docker-compose up --remove-orphans -t 1 --scale nginx=${1:-100}

