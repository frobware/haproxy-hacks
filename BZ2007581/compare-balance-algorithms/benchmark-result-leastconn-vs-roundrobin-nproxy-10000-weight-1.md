| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 738.6 ± 33.6 | 707.6 | 818.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-1` | 745.9 ± 42.1 | 702.4 | 844.2 | 1.01 ± 0.07 |
