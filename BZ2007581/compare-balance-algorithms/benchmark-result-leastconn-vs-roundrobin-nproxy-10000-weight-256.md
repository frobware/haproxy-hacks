| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 755.0 ± 62.6 | 710.6 | 912.3 | 1.01 ± 0.10 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 749.0 ± 45.7 | 711.2 | 846.8 | 1.00 |
