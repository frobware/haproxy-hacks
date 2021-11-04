| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 81.6 ± 17.3 | 62.1 | 131.5 | 1.00 ± 0.28 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 81.4 ± 15.0 | 69.6 | 132.1 | 1.00 |
