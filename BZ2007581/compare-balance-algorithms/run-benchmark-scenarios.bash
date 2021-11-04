#!/usr/bin/env bash

set -e
set -a

ALGORITHMS="leastconn roundrobin" NPROXY=1000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

ALGORITHMS="leastconn roundrobin" NPROXY=1000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

ALGORITHMS="leastconn source" NPROXY=1000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

ALGORITHMS="leastconn source" NPROXY=1000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

ALGORITHMS="leastconn random" NPROXY=1000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

ALGORITHMS="leastconn random" NPROXY=1000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -s none

