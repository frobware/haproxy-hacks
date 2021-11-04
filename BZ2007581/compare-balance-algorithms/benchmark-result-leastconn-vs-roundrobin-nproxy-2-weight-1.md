| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-1` | 2.6 ± 0.8 | 1.8 | 5.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-2-weight-1` | 2.6 ± 1.1 | 1.8 | 8.3 | 1.04 ± 0.54 |
