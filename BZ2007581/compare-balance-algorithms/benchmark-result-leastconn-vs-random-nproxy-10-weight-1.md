| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-1` | 2.9 ± 0.9 | 2.2 | 7.2 | 1.02 ± 0.43 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10-weight-1` | 2.9 ± 0.8 | 2.3 | 7.5 | 1.00 |
