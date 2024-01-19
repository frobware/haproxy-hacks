#!/usr/bin/env bash

set -eu
cd compare-certs
go run main.go "$@"
