| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-2` | 2.6 ± 1.1 | 1.7 | 7.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-2-weight-2` | 2.7 ± 1.0 | 1.8 | 6.8 | 1.03 ± 0.59 |
