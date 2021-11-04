| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 9.2 ± 2.4 | 7.3 | 18.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-100-weight-1` | 9.7 ± 2.5 | 7.2 | 18.1 | 1.05 ± 0.39 |
