# HAProxy Load times

## Configuration: leastconn-vs-random-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 70.8 ± 11.0 | 62.8 | 108.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 83.5 ± 19.0 | 71.8 | 138.1 | 1.18 ± 0.33 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 78.2 ± 18.1 | 64.7 | 116.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-256` | 5414.9 ± 49.2 | 5335.3 | 5497.4 | 69.28 ± 16.08 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 71.6 ± 14.7 | 62.1 | 110.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 72.6 ± 16.3 | 62.2 | 116.0 | 1.01 ± 0.31 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 80.1 ± 22.2 | 62.2 | 122.3 | 1.09 ± 0.39 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 73.3 ± 16.1 | 62.5 | 115.0 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.8 | 1.34 ± 4.68 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.3 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 76.5 ± 18.9 | 62.5 | 114.7 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 82.4 ± 29.1 | 64.2 | 169.3 | 1.08 ± 0.46 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 73.3 ± 14.9 | 62.5 | 110.6 | 1.06 ± 0.25 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-256` | 69.1 ± 7.7 | 63.8 | 105.1 | 1.00 |
