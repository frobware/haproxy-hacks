| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 4.4 ± 2.2 | 2.3 | 13.5 | 1.15 ± 0.82 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10-weight-256` | 3.8 ± 1.9 | 2.2 | 12.4 | 1.00 |
