| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-256` | 2.3 ± 0.9 | 1.6 | 6.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-256` | 2.4 ± 1.0 | 1.7 | 7.7 | 1.06 ± 0.62 |
