#!/usr/bin/env bash

set -euo pipefail

./list-unassociated-eips.sh | awk '{ print $2 }' | paste -sd, -

