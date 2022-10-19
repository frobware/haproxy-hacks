#! /usr/bin/env bash

set -eux

: "${PODMAN:=docker}"

"$PODMAN" run -w /data --rm -it -v "$(pwd)":/data quay.io/amcdermo/mb "$@"


