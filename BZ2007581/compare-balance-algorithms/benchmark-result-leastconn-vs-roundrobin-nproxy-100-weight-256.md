| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 10.2 ± 2.9 | 7.4 | 21.5 | 1.09 ± 0.40 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-100-weight-256` | 9.3 ± 2.2 | 7.3 | 15.3 | 1.00 |
