| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 737.3 ± 18.6 | 716.8 | 771.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-256` | 54823.5 ± 993.0 | 53706.8 | 56970.7 | 74.36 ± 2.31 |
