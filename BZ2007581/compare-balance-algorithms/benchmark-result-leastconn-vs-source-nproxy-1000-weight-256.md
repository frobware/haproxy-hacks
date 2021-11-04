| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 73.3 ± 14.9 | 62.5 | 110.6 | 1.06 ± 0.25 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-256` | 69.1 ± 7.7 | 63.8 | 105.1 | 1.00 |
