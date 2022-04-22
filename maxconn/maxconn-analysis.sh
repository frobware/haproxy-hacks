#!/usr/bin/env bash

set -u
set -o pipefail

t=$(mktemp)
db=$(mktemp)
results_file=${1:?no-results-file}

awk '{ print $1, $2, $3, $4, $5, $6, $7, $8 }' "$results_file" > "$t"

sqlite3 "$db" <<EOF
CREATE TABLE results (
        balance_algorithm TEXT,
        weight INTEGER,
        backends INTEGER,
        threads INTEGER,
        maxconn INTEGER,
        actual_maxconn INTEGER,
        actual_maxsock INTEGER,
        process_size INTEGER);
.separator " "
.import $t results
EOF

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

function a() {
    for balance_algorithm in $(algorithm_values); do
        echo "# $balance_algorithm"
        for weight in $(weight_values); do
            # echo "## $balance_algorithm: weight = $weight"
            for backend_count in $(backend_values); do
                # echo "# $balance_algorithm: weight = $weight, backends = $backend_count"
                for thread_count in $(thread_values); do
                    echo "## weight=$weight backends=$backend_count threads=$thread_count"
                    echo "\`\`\`\`console"
                    echo "\`\`\`sh"
                    sqlite3 $db <<EOF
.headers on
.mode column
.width -1 -1 -1 -1
SELECT maxconn, actual_maxconn as "maxconn (HAProxy)", actual_maxsock as "maxsock (HAProxy)", process_size as "Process Size (KB)"
  FROM results
 WHERE weight == $weight and
       balance_algorithm == "$balance_algorithm" and
       backends == $backend_count and
       threads == $thread_count
ORDER BY actual_maxconn, process_size
EOF
                done
                echo "\`\`\`"
                echo "\`\`\`\`"
            done
        done
    done
}

function b() {
    for balance_algorithm in $(algorithm_values); do
        echo "# $balance_algorithm"
        for weight in $(weight_values); do
            # echo "## $balance_algorithm: weight = $weight"
            for backend_count in $(backend_values); do
                # echo "# $balance_algorithm: weight = $weight, backends = $backend_count"
                for thread_count in $(thread_values); do
                    echo "## algorithm=$balance_algorithm weight=$weight backends=$backend_count threads=$thread_count"
                    printf "\`\`\`\n";
                    sqlite3 $db <<EOF
.headers on
.mode column
.width -1 -1 -1 -1
SELECT maxconn, actual_maxconn as "maxconn (HAProxy)", actual_maxsock as "maxsock (HAProxy)", process_size / 1000 as "Process Size (MB)"
  FROM results
 WHERE weight == $weight and
       balance_algorithm == "$balance_algorithm" and
       backends == $backend_count and
       threads == $thread_count
ORDER BY actual_maxconn, process_size
EOF
                    printf "\`\`\`\n";
                done
            done
        done
    done
}

b
