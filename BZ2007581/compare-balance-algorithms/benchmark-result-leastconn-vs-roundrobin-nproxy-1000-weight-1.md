| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 75.9 ± 13.6 | 64.8 | 106.0 | 1.05 ± 0.26 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 72.4 ± 12.0 | 63.8 | 122.2 | 1.00 |
