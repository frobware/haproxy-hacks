| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 716.3 ± 63.6 | 671.3 | 761.3 | 1.01 ± 0.09 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-256` | 711.7 ± 9.6 | 705.0 | 718.5 | 1.00 |
