| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 9.1 ± 2.4 | 7.3 | 18.5 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-100-weight-256` | 9.2 ± 1.9 | 7.5 | 14.8 | 1.00 ± 0.33 |
