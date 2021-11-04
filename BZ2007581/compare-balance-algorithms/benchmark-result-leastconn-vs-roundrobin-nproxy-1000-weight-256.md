| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 77.0 ± 19.8 | 62.5 | 147.7 | 1.00 ± 0.35 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 76.8 ± 17.7 | 63.2 | 133.5 | 1.00 |
