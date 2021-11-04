# HAProxy Load times

## Configuration: leastconn-random-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.2 ± 0.7 | 1.6 | 6.0 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-1` | 2.5 ± 1.3 | 1.7 | 9.4 | 1.14 ± 0.68 |

## Configuration: leastconn-roundrobin-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.5 ± 0.8 | 1.8 | 5.7 | 1.15 ± 0.49 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 2.2 ± 0.6 | 1.8 | 4.5 | 1.00 |

## Configuration: leastconn-source-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.2 ± 0.6 | 1.7 | 4.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1-weight-1` | 2.3 ± 0.9 | 1.7 | 7.4 | 1.06 ± 0.49 |

## Configuration: leastconn-vs-random-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 774.8 ± 72.5 | 719.4 | 930.2 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-1` | 834.2 ± 11.0 | 821.5 | 858.2 | 1.08 ± 0.10 |

## Configuration: leastconn-vs-random-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 737.3 ± 18.6 | 716.8 | 771.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10000-weight-256` | 54823.5 ± 993.0 | 53706.8 | 56970.7 | 74.36 ± 2.31 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 81.6 ± 17.3 | 62.1 | 131.5 | 1.00 ± 0.28 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-1` | 81.4 ± 15.0 | 69.6 | 132.1 | 1.00 |

## Configuration: leastconn-vs-random-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 78.4 ± 10.9 | 66.9 | 113.2 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1000-weight-256` | 5624.8 ± 65.6 | 5505.6 | 5734.5 | 71.70 ± 10.00 |

## Configuration: leastconn-vs-random-nproxy-100-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 11.5 ± 6.4 | 7.5 | 46.7 | 1.09 ± 0.68 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-100-weight-1` | 10.5 ± 3.0 | 8.2 | 20.0 | 1.00 |

## Configuration: leastconn-vs-random-nproxy-100-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 9.2 ± 2.3 | 7.4 | 17.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-100-weight-256` | 581.6 ± 45.5 | 538.5 | 661.4 | 62.88 ± 16.60 |

## Configuration: leastconn-vs-random-nproxy-10-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-1` | 2.9 ± 0.9 | 2.2 | 7.2 | 1.02 ± 0.43 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10-weight-1` | 2.9 ± 0.8 | 2.3 | 7.5 | 1.00 |

## Configuration: leastconn-vs-random-nproxy-10-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 2.8 ± 0.7 | 2.3 | 7.6 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-10-weight-256` | 69.5 ± 15.8 | 52.5 | 101.7 | 25.08 ± 8.82 |

## Configuration: leastconn-vs-random-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.1 ± 0.6 | 1.7 | 4.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-1` | 2.2 ± 0.6 | 1.7 | 5.0 | 1.03 ± 0.40 |

## Configuration: leastconn-vs-random-nproxy-1-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-256` | 2.3 ± 0.9 | 1.7 | 6.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-256` | 9.4 ± 2.7 | 6.6 | 20.1 | 4.01 ± 1.87 |

## Configuration: leastconn-vs-random-nproxy-1-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-2` | 2.3 ± 0.8 | 1.6 | 7.5 | 1.01 ± 0.50 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-1-weight-2` | 2.3 ± 0.7 | 1.7 | 6.9 | 1.00 |

## Configuration: leastconn-vs-random-nproxy-2-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-1` | 2.4 ± 0.8 | 1.7 | 6.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-2-weight-1` | 2.4 ± 0.8 | 1.7 | 6.9 | 1.01 ± 0.49 |

## Configuration: leastconn-vs-random-nproxy-2-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-2` | 2.6 ± 1.0 | 1.8 | 8.0 | 1.10 ± 0.54 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-random-nproxy-2-weight-2` | 2.4 ± 0.7 | 1.8 | 6.2 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 786.4 ± 105.4 | 728.4 | 1080.5 | 1.01 ± 0.16 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-1` | 775.8 ± 60.1 | 729.9 | 905.9 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 763.8 ± 33.2 | 712.8 | 825.2 | 1.00 ± 0.07 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10000-weight-256` | 763.3 ± 46.3 | 710.2 | 846.1 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 75.2 ± 15.0 | 62.5 | 110.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-1` | 77.5 ± 15.6 | 63.6 | 120.1 | 1.03 ± 0.29 |

## Configuration: leastconn-vs-roundrobin-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 77.0 ± 19.8 | 62.5 | 147.7 | 1.00 ± 0.35 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1000-weight-256` | 76.8 ± 17.7 | 63.2 | 133.5 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-100-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 9.2 ± 2.4 | 7.3 | 18.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-100-weight-1` | 9.7 ± 2.5 | 7.2 | 18.1 | 1.05 ± 0.39 |

## Configuration: leastconn-vs-roundrobin-nproxy-100-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 10.2 ± 2.9 | 7.4 | 21.5 | 1.09 ± 0.40 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-100-weight-256` | 9.3 ± 2.2 | 7.3 | 15.3 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-10-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-1` | 3.1 ± 1.1 | 2.2 | 6.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10-weight-1` | 3.4 ± 1.3 | 2.3 | 8.0 | 1.10 ± 0.56 |

## Configuration: leastconn-vs-roundrobin-nproxy-10-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 4.4 ± 2.2 | 2.3 | 13.5 | 1.15 ± 0.82 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-10-weight-256` | 3.8 ± 1.9 | 2.2 | 12.4 | 1.00 |

## Configuration: leastconn-vs-roundrobin-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.3 ± 0.9 | 1.7 | 7.2 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-1` | 2.4 ± 1.0 | 1.7 | 5.9 | 1.06 ± 0.57 |

## Configuration: leastconn-vs-roundrobin-nproxy-1-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-256` | 2.3 ± 0.9 | 1.6 | 6.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-256` | 2.4 ± 1.0 | 1.7 | 7.7 | 1.06 ± 0.62 |

## Configuration: leastconn-vs-roundrobin-nproxy-1-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-2` | 2.2 ± 0.7 | 1.7 | 4.6 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-1-weight-2` | 2.3 ± 0.7 | 1.7 | 5.8 | 1.01 ± 0.45 |

## Configuration: leastconn-vs-roundrobin-nproxy-2-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-1` | 2.6 ± 0.8 | 1.8 | 5.1 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-2-weight-1` | 2.6 ± 1.1 | 1.8 | 8.3 | 1.04 ± 0.54 |

## Configuration: leastconn-vs-roundrobin-nproxy-2-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-2` | 2.6 ± 1.1 | 1.7 | 7.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-roundrobin-nproxy-2-weight-2` | 2.7 ± 1.0 | 1.8 | 6.8 | 1.03 ± 0.59 |

## Configuration: leastconn-vs-source-nproxy-10000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-1` | 753.8 ± 20.8 | 729.9 | 801.6 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-1` | 760.2 ± 24.7 | 730.4 | 810.5 | 1.01 ± 0.04 |

## Configuration: leastconn-vs-source-nproxy-10000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10000-weight-256` | 738.4 ± 19.5 | 721.0 | 778.9 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10000-weight-256` | 749.3 ± 25.3 | 726.6 | 811.2 | 1.01 ± 0.04 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-1` | 89.5 ± 42.6 | 64.8 | 271.4 | 1.16 ± 0.61 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-1` | 77.2 ± 16.9 | 63.2 | 126.6 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-1000-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1000-weight-256` | 74.7 ± 13.2 | 63.1 | 115.8 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1000-weight-256` | 78.1 ± 13.6 | 64.2 | 108.0 | 1.04 ± 0.26 |

## Configuration: leastconn-vs-source-nproxy-100-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-1` | 10.1 ± 3.1 | 7.4 | 23.3 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-100-weight-1` | 10.2 ± 2.6 | 7.4 | 17.7 | 1.01 ± 0.41 |

## Configuration: leastconn-vs-source-nproxy-100-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-100-weight-256` | 9.1 ± 2.4 | 7.3 | 18.5 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-100-weight-256` | 9.2 ± 1.9 | 7.5 | 14.8 | 1.00 ± 0.33 |

## Configuration: leastconn-vs-source-nproxy-10-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-1` | 3.2 ± 1.1 | 2.3 | 7.5 | 1.12 ± 0.49 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10-weight-1` | 2.9 ± 0.8 | 2.2 | 6.0 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-10-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-10-weight-256` | 3.1 ± 1.8 | 2.1 | 16.4 | 1.07 ± 0.69 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-10-weight-256` | 2.9 ± 0.8 | 2.2 | 6.0 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-1-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-1` | 2.3 ± 0.9 | 1.6 | 7.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1-weight-1` | 2.3 ± 0.9 | 1.6 | 6.9 | 1.02 ± 0.55 |

## Configuration: leastconn-vs-source-nproxy-1-weight-256
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-256` | 2.3 ± 0.8 | 1.7 | 5.4 | 1.00 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1-weight-256` | 2.4 ± 0.9 | 1.7 | 6.4 | 1.02 ± 0.50 |

## Configuration: leastconn-vs-source-nproxy-1-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-1-weight-2` | 2.3 ± 0.7 | 1.7 | 5.8 | 1.02 ± 0.44 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-1-weight-2` | 2.3 ± 0.7 | 1.7 | 6.5 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-2-weight-1
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-1` | 3.9 ± 6.3 | 1.4 | 28.3 | 2.08 ± 3.43 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-2-weight-1` | 1.9 ± 0.7 | 1.3 | 5.0 | 1.00 |

## Configuration: leastconn-vs-source-nproxy-2-weight-2
| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-leastconn-nproxy-2-weight-2` | 2.4 ± 0.8 | 1.7 | 6.1 | 1.02 ± 0.46 |
| `haproxy -c -f haproxy.config -C benchmark-config-algorithm-source-nproxy-2-weight-2` | 2.3 ± 0.7 | 1.7 | 5.5 | 1.00 |
