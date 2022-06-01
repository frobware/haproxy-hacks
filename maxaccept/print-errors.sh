#!/usr/bin/env bash

set -eu

response_file=${1:?no-responses.csv-file-specified}
db=$(mktemp)

sqlite3 "$db" <<EOF
CREATE TABLE results (
    start_request     INTEGER,
    delay             INTEGER,
    status            INTEGER,
    written           INTEGER,
    read              INTEGER,
    method_and_url    TEXT,
    thread_id         INTEGER,
    conn_id           INTEGER,
    conns             INTEGER,
    reqs              INTEGER,
    start             INTEGER,
    socket_writeable  INTEGER,
    conn_est          INTEGER,
    tls_reuse         INTEGER,
    err               TEXT DEFAULT NULL
);
.separator ","
.import $response_file results
EOF

echo "SQLiteDB: $db"
echo "$(cat $response_file | wc -l) responses"

echo
sqlite3 "$db" <<EOF
.headers on
.mode column

SELECT count(status) AS "count(status != 200)"
FROM   results
WHERE  status != 200;

SELECT DISTINCT err
FROM            results
WHERE           err IS NOT "";
EOF
