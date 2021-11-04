#!/usr/bin/env bash

set -e
set -a

ALGORITHMS="leastconn roundrobin" NPROXY=10000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

ALGORITHMS="leastconn roundrobin" NPROXY=10000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

ALGORITHMS="leastconn source" NPROXY=10000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

ALGORITHMS="leastconn source" NPROXY=10000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

ALGORITHMS="leastconn random" NPROXY=10000 WEIGHTS="1"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

ALGORITHMS="leastconn random" NPROXY=10000 WEIGHTS="256"
./benchmark.bash --export-markdown benchmark-result-${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md -M=2

