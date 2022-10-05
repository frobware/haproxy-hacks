#!/usr/bin/env bash

set -eux

C=1 ./generate-pods.sh
C=1 ./generate-services.sh
C=1 ./generate-routes.sh
