| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 9.2 ± 2.3 | 7.4 | 17.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-100-weight-256` | 581.6 ± 45.5 | 538.5 | 661.4 | 62.88 ± 16.60 |
