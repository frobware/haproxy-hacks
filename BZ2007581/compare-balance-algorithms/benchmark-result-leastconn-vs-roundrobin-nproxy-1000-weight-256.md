| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 98.8 ± 44.1 | 66.3 | 258.8 | 1.38 ± 0.64 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 71.7 ± 9.2 | 62.7 | 101.3 | 1.00 |
