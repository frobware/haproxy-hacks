| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-1` | 3.1 ± 1.1 | 2.2 | 6.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10-weight-1` | 3.4 ± 1.3 | 2.3 | 8.0 | 1.10 ± 0.56 |
