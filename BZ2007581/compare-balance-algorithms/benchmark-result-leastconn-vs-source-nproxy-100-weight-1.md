| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 10.1 ± 3.1 | 7.4 | 23.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-100-weight-1` | 10.2 ± 2.6 | 7.4 | 17.7 | 1.01 ± 0.41 |
