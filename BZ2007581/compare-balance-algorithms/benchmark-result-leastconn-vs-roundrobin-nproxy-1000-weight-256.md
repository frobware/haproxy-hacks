| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 80.1 ± 22.2 | 62.2 | 122.3 | 1.09 ± 0.39 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 73.3 ± 16.1 | 62.5 | 115.0 | 1.00 |
