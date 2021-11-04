# HAProxy Load times

## Configuration: leastconn-vs-random-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 721.5 ± 52.9 | 684.1 | 758.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-1` | 762.5 ± 3.3 | 760.2 | 764.8 | 1.06 ± 0.08 |

## Configuration: leastconn-vs-random-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 815.1 ± 125.8 | 726.1 | 904.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-256` | 52979.9 ± 147.0 | 52875.9 | 53083.8 | 65.00 ± 10.04 |

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

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 669.6 ± 6.5 | 665.0 | 674.2 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-1` | 672.6 ± 11.6 | 664.4 | 680.8 | 1.00 ± 0.02 |

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 818.5 ± 11.3 | 810.5 | 826.5 | 1.06 ± 0.18 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 769.9 ± 131.3 | 677.0 | 862.7 | 1.00 |

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

## Configuration: leastconn-vs-source-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 677.1 ± 7.1 | 672.0 | 682.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-1` | 710.2 ± 48.0 | 676.3 | 744.2 | 1.05 ± 0.07 |

## Configuration: leastconn-vs-source-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 716.3 ± 63.6 | 671.3 | 761.3 | 1.01 ± 0.09 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-256` | 711.7 ± 9.6 | 705.0 | 718.5 | 1.00 |

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
