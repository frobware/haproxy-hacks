| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.8 | 1.34 ± 4.68 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.3 | 1.00 |
