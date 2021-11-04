#!/usr/bin/env bash

ALGORITHMS="leastconn roundrobin" NPROXY=1000 WEIGHTS="1" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

ALGORITHMS="leastconn roundrobin" NPROXY=1000 WEIGHTS="256" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

ALGORITHMS="leastconn source" NPROXY=1000 WEIGHTS="1" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

ALGORITHMS="leastconn source" NPROXY=1000 WEIGHTS="256" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

ALGORITHMS="leastconn random" NPROXY=1000 WEIGHTS="1" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

ALGORITHMS="leastconn random" NPROXY=1000 WEIGHTS="256" \
	  ./benchmark.sh --export-markdown benchmark-result:${ALGORITHMS// /-vs-}-nproxy-${NPROXY}-weight-${WEIGHTS}.md

