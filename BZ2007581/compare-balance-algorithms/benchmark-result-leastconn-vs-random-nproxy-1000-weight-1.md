| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 84.0 ± 14.3 | 66.4 | 114.0 | 1.02 ± 0.20 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 82.7 ± 8.7 | 73.5 | 118.4 | 1.00 |
