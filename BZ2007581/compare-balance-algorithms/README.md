# HAProxy Load times

## Configuration: leastconn-vs-random-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 717.4 ± 8.6 | 703.9 | 729.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-1` | 812.3 ± 17.5 | 787.6 | 837.1 | 1.13 ± 0.03 |

## Configuration: leastconn-vs-random-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 735.9 ± 41.5 | 699.0 | 825.7 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-256` | 55182.3 ± 599.4 | 54342.8 | 55823.9 | 74.98 ± 4.31 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 84.0 ± 14.3 | 66.4 | 114.0 | 1.02 ± 0.20 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 82.7 ± 8.7 | 73.5 | 118.4 | 1.00 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 75.9 ± 13.9 | 62.1 | 118.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-256` | 5655.1 ± 231.3 | 5395.3 | 6044.0 | 74.51 ± 13.98 |

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 738.6 ± 33.6 | 707.6 | 818.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-1` | 745.9 ± 42.1 | 702.4 | 844.2 | 1.01 ± 0.07 |

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 755.0 ± 62.6 | 710.6 | 912.3 | 1.01 ± 0.10 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 749.0 ± 45.7 | 711.2 | 846.8 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 75.9 ± 13.6 | 64.8 | 106.0 | 1.05 ± 0.26 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 72.4 ± 12.0 | 63.8 | 122.2 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 98.8 ± 44.1 | 66.3 | 258.8 | 1.38 ± 0.64 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 71.7 ± 9.2 | 62.7 | 101.3 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 738.6 ± 32.9 | 710.1 | 805.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-1` | 749.5 ± 34.2 | 721.6 | 839.0 | 1.01 ± 0.06 |

## Configuration: leastconn-vs-source-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 729.7 ± 20.1 | 716.0 | 784.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-256` | 745.6 ± 30.0 | 722.9 | 827.5 | 1.02 ± 0.05 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 73.7 ± 11.9 | 63.0 | 106.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 77.6 ± 12.7 | 64.2 | 112.1 | 1.05 ± 0.24 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 75.2 ± 15.3 | 62.2 | 117.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-256` | 77.1 ± 11.6 | 65.5 | 105.1 | 1.03 ± 0.26 |
