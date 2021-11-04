| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f benchmark-config:algorithm-leastconn-nproxy-1000-weight-256/haproxy.config` | 151.8 ± 20.7 | 137.8 | 211.9 | 1.00 |
| `haproxy -c -f benchmark-config:algorithm-random-nproxy-1000-weight-256/haproxy.config` | 5415.7 ± 88.4 | 5230.4 | 5546.7 | 35.68 ± 4.90 |
