| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 73.7 ± 11.9 | 63.0 | 106.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 77.6 ± 12.7 | 64.2 | 112.1 | 1.05 ± 0.24 |
