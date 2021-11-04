| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 2.8 ± 0.7 | 2.3 | 7.6 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10-weight-256` | 69.5 ± 15.8 | 52.5 | 101.7 | 25.08 ± 8.82 |
