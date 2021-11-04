| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 11.5 ± 6.4 | 7.5 | 46.7 | 1.09 ± 0.68 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-100-weight-1` | 10.5 ± 3.0 | 8.2 | 20.0 | 1.00 |
