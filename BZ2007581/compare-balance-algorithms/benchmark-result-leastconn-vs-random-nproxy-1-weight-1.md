| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.1 ± 0.6 | 1.7 | 4.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-1` | 2.2 ± 0.6 | 1.7 | 5.0 | 1.03 ± 0.40 |
