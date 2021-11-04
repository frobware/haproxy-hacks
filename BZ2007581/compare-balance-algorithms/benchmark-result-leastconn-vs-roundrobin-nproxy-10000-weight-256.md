| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 763.8 ± 33.2 | 712.8 | 825.2 | 1.00 ± 0.07 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 763.3 ± 46.3 | 710.2 | 846.1 | 1.00 |
