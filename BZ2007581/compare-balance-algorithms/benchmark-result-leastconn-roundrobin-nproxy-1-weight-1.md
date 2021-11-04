| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.5 ± 0.8 | 1.8 | 5.7 | 1.15 ± 0.49 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 2.2 ± 0.6 | 1.8 | 4.5 | 1.00 |
