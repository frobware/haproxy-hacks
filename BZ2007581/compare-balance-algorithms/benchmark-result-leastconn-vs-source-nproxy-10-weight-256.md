| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 3.1 ± 1.8 | 2.1 | 16.4 | 1.07 ± 0.69 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10-weight-256` | 2.9 ± 0.8 | 2.2 | 6.0 | 1.00 |
