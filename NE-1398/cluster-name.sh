#!/usr/bin/env bash

set -euo pipefail

echo "$(oc get infrastructure cluster -o jsonpath='{.status.infrastructureName}')"
