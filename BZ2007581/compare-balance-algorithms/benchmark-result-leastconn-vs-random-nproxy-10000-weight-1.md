| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 721.5 ± 52.9 | 684.1 | 758.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-1` | 762.5 ± 3.3 | 760.2 | 764.8 | 1.06 ± 0.08 |
