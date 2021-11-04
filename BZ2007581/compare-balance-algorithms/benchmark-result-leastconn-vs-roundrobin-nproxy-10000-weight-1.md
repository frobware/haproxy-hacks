| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 786.4 ± 105.4 | 728.4 | 1080.5 | 1.01 ± 0.16 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-1` | 775.8 ± 60.1 | 729.9 | 905.9 | 1.00 |
