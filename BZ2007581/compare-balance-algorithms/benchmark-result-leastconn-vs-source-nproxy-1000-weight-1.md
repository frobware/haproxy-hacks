| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 76.5 ± 18.9 | 62.5 | 114.7 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 82.4 ± 29.1 | 64.2 | 169.3 | 1.08 ± 0.46 |
