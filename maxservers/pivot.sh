#!/usr/bin/env bash

set -u
set -o pipefail

t=$(mktemp)
db=$(mktemp)
results_file=${1:?no-results-file}

awk '{ print $1, $2, $3, $4, $5, $6, $7, $8 }' "$results_file" > "$t"

sqlite3 "$db" <<EOF
CREATE TABLE results (
        weight INTEGER,
        balance_algorithm TEXT,
        backends INTEGER,
        servers INTEGER,
        threads INTEGER,
        process_size_in_kb INTEGER,
        process_size_in_mb INTEGER,
        use_server_template INTEGER);
.separator " "
.import $t results
EOF

echo "# [NE-1690 - Analyse Memory Impact of Pre-Allocated Server Slots for Different Numbers of Routes](https://issues.redhat.com/browse/NE-1690)"

a=$(ocp-haproxy-2.8.5 -v | head -1)
echo "$a"
echo
echo "Column headers:"
echo "- ST=<0|1> (server-template disabled=0, enabled=1)"
echo "- Tn (Number of Threads)"
echo "- RSS (Memory usage in MB)"

echo "
I have been collecting memory usage data for HAProxy's server-template
feature, where the range is 0..N - this is the ST=1 column.
Additionally, I've collected data where server lines are explicitly
expanded into the haproxy.config for 0..N - this is the ST=0 column.

The memory usage is broadly similar in both cases. However, if you
don't use the server-template, you avoid incurring the memory cost
upfront. When every slot in the server-template is used, it equates to
using the same number of server lines in a backend.

The runtime API provides the ability to add and delete servers
dynamically. Given this API capability, there is no compelling reason
to use the server-template feature.
"

generate_pivot_table() {
    local algorithm=$1

    printf "\`\`\`\n"
    sqlite3 $db <<EOF
.headers on
.mode column
.width -9 -8 -11 -11 -12 -12
WITH ranked_results AS (
  SELECT
    backends,
    servers,
    max(case when threads = 4 and use_server_template = 0 then process_size_in_mb end) as "ST=0 T4 RSS",
    max(case when threads = 4 and use_server_template = 1 then process_size_in_mb end) as "ST=1 T4 RSS",
    max(case when threads = 64 and use_server_template = 0 then process_size_in_mb end) as "ST=0 T64 RSS",
    max(case when threads = 64 and use_server_template = 1 then process_size_in_mb end) as "ST=1 T64 RSS",
    ROW_NUMBER() OVER (PARTITION BY backends ORDER BY backends, servers) as rn
  FROM results
  WHERE balance_algorithm = '$algorithm'
  GROUP BY backends, servers
)
SELECT
  CASE WHEN rn = 1 THEN CAST(backends AS TEXT) ELSE '' END as "#backends",
  servers as "#servers",
  "ST=0 T4 RSS",
  "ST=1 T4 RSS",
  "ST=0 T64 RSS",
  "ST=1 T64 RSS"
FROM ranked_results
ORDER BY backends, servers;
EOF
    printf "\`\`\`\n"
}

for algorithm in $(sqlite3 $db "select distinct(balance_algorithm) from results"); do
    echo "## Algorithm=$algorithm, Weight=1 maxconn=50000"
    generate_pivot_table $algorithm
done
