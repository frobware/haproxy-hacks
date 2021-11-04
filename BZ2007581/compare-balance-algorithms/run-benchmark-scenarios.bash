#!/usr/bin/env bash

set -eu

algorithm_combos=("leastconn,roundrobin"
		  "leastconn,source"
		  "leastconn,random")

for nproxy in 1000 10000; do
    for weight in 1 256; do
	for algorithm_combo in "${algorithm_combos[@]}"; do
	    hyperfine \
		-L nproxy $nproxy \
		-L weight $weight \
		-L algorithm "$algorithm_combo" \
		--prepare './generate-haproxy-config.pl --balance-algorithm={algorithm} --nproxy={nproxy} --weight={weight} --output-dir=benchmark-config-algorithm-{algorithm}-nproxy-{nproxy}-weight-{weight}' \
		--export-markdown "benchmark-result-${algorithm_combo/,/-vs-}-nproxy-${nproxy}-weight-${weight}.md" \
		"$@" \
		'haproxy -c -f haproxy.config -C benchmark-config-algorithm-{algorithm}-nproxy-{nproxy}-weight-{weight}'
	done
    done
done
