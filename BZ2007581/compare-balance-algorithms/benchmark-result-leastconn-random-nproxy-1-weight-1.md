| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.2 ± 0.7 | 1.6 | 6.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-1` | 2.5 ± 1.3 | 1.7 | 9.4 | 1.14 ± 0.68 |
