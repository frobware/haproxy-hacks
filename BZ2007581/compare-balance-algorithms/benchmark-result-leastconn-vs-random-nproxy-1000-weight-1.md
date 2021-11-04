| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 70.8 ± 11.0 | 62.8 | 108.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 83.5 ± 19.0 | 71.8 | 138.1 | 1.18 ± 0.33 |
