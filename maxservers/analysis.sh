#!/usr/bin/env bash

set -u
set -o pipefail

t=$(mktemp)
db=$(mktemp)
results_file=${1:?no-results-file}

awk '{ print $1, $2, $3, $4, $5, $6, $7 }' "$results_file" > "$t"

sqlite3 "$db" <<EOF
CREATE TABLE results (
        weight INTEGER,
        balance_algorithm TEXT,
        backends INTEGER,
        servers INTEGER,
        threads INTEGER,
        process_size_in_kb INTEGER,
        process_size_in_mb INTEGER);
.separator " "
.import $t results
EOF

echo "$db";

function weight_values() {
    echo $(sqlite3 $db "select distinct(weight) from results")
}

function algorithm_values() {
    echo $(sqlite3 $db "select distinct(balance_algorithm) from results")
}

function backend_values() {
    echo $(sqlite3 $db "select distinct(backends) from results")
}

function thread_values() {
    echo $(sqlite3 $db "select distinct(threads) from results")
}

for algorithm in $(algorithm_values); do
    # echo "# $algorithm"
    for weight in $(weight_values); do
        # echo "## $algorithm: weight = $weight"
        for backend_count in $(backend_values); do
            # echo "# $algorithm: weight = $weight, backends = $backend_count"
            for thread_count in $(thread_values); do
                echo "## algorithm=$algorithm weight=$weight backends=$backend_count threads=$thread_count"
                printf "\`\`\`\n";
                sqlite3 $db <<EOF
.headers on
.mode column
.width -10 -18
SELECT servers as "#servers", process_size_in_kb / 1024 as "Process Size (MB)"
  FROM results
 WHERE weight == $weight and
       balance_algorithm == '$algorithm' and
       backends == $backend_count and
       threads == $thread_count
ORDER BY servers, process_size_in_kb
EOF
                printf "\`\`\`\n";
            done
        done
    done
done
