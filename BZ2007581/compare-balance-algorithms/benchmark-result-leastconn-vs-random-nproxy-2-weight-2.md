| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-2` | 2.6 ± 1.0 | 1.8 | 8.0 | 1.10 ± 0.54 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-2-weight-2` | 2.4 ± 0.7 | 1.8 | 6.2 | 1.00 |
