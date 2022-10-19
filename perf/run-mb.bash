#! /usr/bin/env bash

set -eu

: "${DOCKER:=podman}"

"$DOCKER" run --rm -it -v "$(pwd)":/data quay.io/amcdermo/mb -i "${1:-requests.json}" -o "${2:-/dev/stdout}"

