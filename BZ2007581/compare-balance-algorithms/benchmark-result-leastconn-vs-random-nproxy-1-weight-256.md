| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-256` | 2.3 ± 0.9 | 1.7 | 6.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-256` | 9.4 ± 2.7 | 6.6 | 20.1 | 4.01 ± 1.87 |
