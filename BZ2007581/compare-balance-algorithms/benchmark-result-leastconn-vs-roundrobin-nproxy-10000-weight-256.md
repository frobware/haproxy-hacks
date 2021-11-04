| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 818.5 ± 11.3 | 810.5 | 826.5 | 1.06 ± 0.18 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 769.9 ± 131.3 | 677.0 | 862.7 | 1.00 |
