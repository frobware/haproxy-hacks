# Comparison of haproxy load times & balance algorithms

## Configuration: leastconn-vs-random-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 0.1 ± 0.1 | 0.0 | 0.9 | 1.00 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 0.2 ± 0.2 | 0.0 | 1.4 | 1.64 ± 3.38 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 0.2 ± 0.3 | 0.0 | 2.1 | 1.22 ± 3.22 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-256` | 0.1 ± 0.2 | 0.0 | 1.6 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 0.1 ± 0.2 | 0.0 | 1.3 | 1.20 ± 2.67 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 0.1 ± 0.2 | 0.0 | 0.7 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 0.1 ± 0.2 | 0.0 | 0.8 | 1.08 ± 2.37 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 0.1 ± 0.2 | 0.0 | 1.5 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.8 | 1.34 ± 4.68 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 0.1 ± 0.2 | 0.0 | 1.3 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 0.1 ± 0.2 | 0.0 | 1.1 | 1.06 ± 2.32 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 0.1 ± 0.2 | 0.0 | 1.1 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 0.2 ± 0.3 | 0.0 | 3.4 | 1.00 |
| `echo haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-256` | 0.2 ± 0.2 | 0.0 | 1.1 | 1.12 ± 2.46 |
