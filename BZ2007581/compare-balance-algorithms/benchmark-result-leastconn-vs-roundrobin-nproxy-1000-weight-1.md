| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 71.6 ± 14.7 | 62.1 | 110.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 72.6 ± 16.3 | 62.2 | 116.0 | 1.01 ± 0.31 |
