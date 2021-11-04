| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-2` | 2.2 ± 0.7 | 1.7 | 4.6 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-2` | 2.3 ± 0.7 | 1.7 | 5.8 | 1.01 ± 0.45 |
