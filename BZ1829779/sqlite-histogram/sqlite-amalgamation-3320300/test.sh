#!/usr/bin/env bash

./sqlite3 <<EOF
CREATE TABLE results (time_total REAL);
INSERT INTO results values(100);
.load ./histograms
SELECT * FROM Histo("results", "time_total", 10, 0, 100);
EOF
