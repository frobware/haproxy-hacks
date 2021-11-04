| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 738.4 ± 19.5 | 721.0 | 778.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-256` | 749.3 ± 25.3 | 726.6 | 811.2 | 1.01 ± 0.04 |
