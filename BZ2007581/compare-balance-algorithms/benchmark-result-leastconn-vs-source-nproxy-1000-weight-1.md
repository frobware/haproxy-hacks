| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 89.5 ± 42.6 | 64.8 | 271.4 | 1.16 ± 0.61 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 77.2 ± 16.9 | 63.2 | 126.6 | 1.00 |
