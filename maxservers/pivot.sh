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

echo "# [NE-1690 -Analyse Memory Impact of Pre-Allocated Server Slots for Different Numbers of Routes](https://issues.redhat.com/browse/NE-1690)"

# Generate pivot table for each combination of algorithm, weight
for algorithm in $(sqlite3 $db "select distinct(balance_algorithm) from results"); do
    for weight in $(sqlite3 $db "select distinct(weight) from results"); do
        echo "## Algorithm=$algorithm, Weight=$weight maxconn=50000 (T=#Threads B=#Backends)"
        printf "\`\`\`\n";
        sqlite3 $db <<EOF
.headers on
.mode column
.width -20 -16 -17 -18 -17 -18 -19
SELECT servers as "#Servers per Backend",
       max(case when threads = 4 and backends = 100 then process_size_in_mb end) as "T4 B100 RSS (MB)",
       max(case when threads = 4 and backends = 1000 then process_size_in_mb end) as "T4 B1000 RSS (MB)",
       max(case when threads = 4 and backends = 10000 then process_size_in_mb end) as "T4 B10000 RSS (MB)",
       max(case when threads = 64 and backends = 100 then process_size_in_mb end) as "T64 B100 RSS (MB)",
       max(case when threads = 64 and backends = 1000 then process_size_in_mb end) as "T64 B1000 RSS (MB)",
       max(case when threads = 64 and backends = 10000 then process_size_in_mb end) as "T64 B10000 RSS (MB)"
  FROM results
  WHERE weight = $weight AND
        balance_algorithm = '$algorithm'
  GROUP BY servers
  ORDER BY servers;
EOF
        printf "\`\`\`\n";
    done
done
