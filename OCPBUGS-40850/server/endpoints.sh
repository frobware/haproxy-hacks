#!/usr/bin/env bash

set -eu

# if ! curl -sS "http://localhost:${1:-1051}/discovery"; then
#     exit 1
# fi

./get.sh $1 /discovery
