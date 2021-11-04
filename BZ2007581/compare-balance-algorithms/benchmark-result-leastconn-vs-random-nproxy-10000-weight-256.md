| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 815.1 ± 125.8 | 726.1 | 904.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-256` | 52979.9 ± 147.0 | 52875.9 | 53083.8 | 65.00 ± 10.04 |
