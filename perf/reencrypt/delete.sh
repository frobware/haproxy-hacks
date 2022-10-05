#!/usr/bin/env bash

set -eux

./generate-pods.sh
./generate-services.sh
./generate-routes.sh
