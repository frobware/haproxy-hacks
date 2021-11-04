| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 677.1 ± 7.1 | 672.0 | 682.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-1` | 710.2 ± 48.0 | 676.3 | 744.2 | 1.05 ± 0.07 |
