#!/usr/bin/env perl

use strict;
use warnings;

sub generate_table {
    my ($termination_type, $metric, $values) = @_;

    print "** $termination_type\n";

    print "| Release | $metric | % Diff from 4.13.8 |\n";
    print "|-|-|-|\n";
    foreach my $hash (@$values) {
	foreach my $release (keys %$hash) {
	    my $val = $hash->{$release};
	    printf "| %s | %.0f | |\n", $release, $val;
	}
    }
    print q!#+TBLFM: $3=(($2 - @2$2) / @2$2) * 100;%.2f!;
    print "\n";
}

  # "version": "wip@7459238ff534498ab5a26cd1a45684cdd327f84a",
  # "platform": "None",
  # "clusterType": "self-managed",
  # "ocpVersion": "4.14.0-0.nightly-2023-08-11-055332",
  # "ocpMajorVersion": "4.14",
  # "k8sVersion": "v1.27.4+deb2c60",
  # "masterNodesType": "",
  # "workerNodesType": "",
  # "masterNodesCount": 3,
  # "infraNodesType": "",
  # "workerNodesCount": 3,
  # "infraNodesCount": 2,
  # "otherNodesCount": 0,
  # "totalNodes": 8,
  # "sdnType": "OVNKubernetes",

# ocp 4.14 haproxy-2.2
# % ag overall y
# 478:time="2023-09-08T13:20:09+01:00" level=info msg="Overall for http samples=2 16243 requests/s avgLatency=127ms P99Latency=276ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T13:30:11+01:00" level=info msg="Overall for edge samples=2 15961 requests/s avgLatency=125ms P99Latency=253ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T13:40:13+01:00" level=info msg="Overall for reencrypt samples=2 16177 requests/s avgLatency=124ms P99Latency=248ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T13:50:15+01:00" level=info msg="Overall for passthrough samples=2 14652 requests/s avgLatency=133ms P99Latency=263ms sum(timeouts)=0 sum(httperrors)=0"

# ocp 4.13 haproxy-2.2
# % ag overall ~/x
# 478:time="2023-09-08T14:00:53+01:00" level=info msg="Overall for http samples=2 29239 requests/s avgLatency=71ms P99Latency=267ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T14:10:56+01:00" level=info msg="Overall for edge samples=2 25749 requests/s avgLatency=86ms P99Latency=354ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T14:20:58+01:00" level=info msg="Overall for reencrypt samples=2 29107 requests/s avgLatency=70ms P99Latency=250ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T14:31:00+01:00" level=info msg="Overall for passthrough samples=2 23160 requests/s avgLatency=87ms P99Latency=298ms sum(timeouts)=0 sum(httperrors)=0"

# ocp 4.14 haproxy-2.6
# % ag overall ~/y
# 478:time="2023-09-08T14:48:38+01:00" level=info msg="Overall for http samples=2 16116 requests/s avgLatency=128ms P99Latency=280ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T14:58:39+01:00" level=info msg="Overall for edge samples=2 15846 requests/s avgLatency=126ms P99Latency=254ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T15:08:41+01:00" level=info msg="Overall for reencrypt samples=2 15964 requests/s avgLatency=125ms P99Latency=251ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T15:18:43+01:00" level=info msg="Overall for passthrough samples=2 14742 requests/s avgLatency=132ms P99Latency=260ms sum(timeouts)=0 sum(httperrors)=0"

# % ag overall ~/x
# 478:time="2023-09-08T16:12:06+01:00" level=info msg="Overall for http samples=2 30098 requests/s avgLatency=69ms P99Latency=232ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T16:22:08+01:00" level=info msg="Overall for edge samples=2 26049 requests/s avgLatency=85ms P99Latency=350ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T16:32:11+01:00" level=info msg="Overall for reencrypt samples=2 27815 requests/s avgLatency=74ms P99Latency=235ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T16:42:14+01:00" level=info msg="Overall for passthrough samples=2 23475 requests/s avgLatency=85ms P99Latency=292ms sum(timeouts)=0 sum(httperrors)=0"

print "* Requests/s\n";
{
    my @http_rps = (
	{'4.13.8'     => 29329},
	{'4.13.8 haproxy-2.6'     => 30098},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 16243},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 16116},
	);

    my @edge_rps = (
	{'4.13.8'     => 25749},
	{'4.13.8 haproxy-2.6'     => 26049},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 15961},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 15846},
	);

    my @reencrypt_rps = (
	{'4.13.8'     =>  29107},
	{'4.13.8 haproxy-2.6'     => 27815},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 16177 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 15964},
	);

    my @passthrough_rps = (
	{'4.13.8'     =>  23160},
	{'4.13.8 haproxy-2.6'     => 23475},
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 14652 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 14742},
	);

    generate_table("http", "request/s (x1000)", \@http_rps);
    print "\n";
    generate_table("edge", "request/s (x1000)", \@edge_rps);
    print "\n";
    generate_table("reencrypt", "reencrypt/s (x1000)", \@reencrypt_rps);
    print "\n";
    generate_table("passthrough", "request/s (x1000)", \@passthrough_rps);
    print "\n";
}

# ocp 4.14 haproxy-2.2
# % ag overall y
# 478:time="2023-09-08T13:20:09+01:00" level=info msg="Overall for http samples=2 16243 requests/s avgLatency=127ms P99Latency=276ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T13:30:11+01:00" level=info msg="Overall for edge samples=2 15961 requests/s avgLatency=125ms P99Latency=253ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T13:40:13+01:00" level=info msg="Overall for reencrypt samples=2 16177 requests/s avgLatency=124ms P99Latency=248ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T13:50:15+01:00" level=info msg="Overall for passthrough samples=2 14652 requests/s avgLatency=133ms P99Latency=263ms sum(timeouts)=0 sum(httperrors)=0"

# ocp 4.13 haproxy-2.2
# % ag overall ~/x
# 478:time="2023-09-08T14:00:53+01:00" level=info msg="Overall for http samples=2 29239 requests/s avgLatency=71ms P99Latency=267ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T14:10:56+01:00" level=info msg="Overall for edge samples=2 25749 requests/s avgLatency=86ms P99Latency=354ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T14:20:58+01:00" level=info msg="Overall for reencrypt samples=2 29107 requests/s avgLatency=70ms P99Latency=250ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T14:31:00+01:00" level=info msg="Overall for passthrough samples=2 23160 requests/s avgLatency=87ms P99Latency=298ms sum(timeouts)=0 sum(httperrors)=0"

# ocp 4.14 haproxy-2.6
# % ag overall ~/y
# 478:time="2023-09-08T14:48:38+01:00" level=info msg="Overall for http samples=2 16116 requests/s avgLatency=128ms P99Latency=280ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T14:58:39+01:00" level=info msg="Overall for edge samples=2 15846 requests/s avgLatency=126ms P99Latency=254ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T15:08:41+01:00" level=info msg="Overall for reencrypt samples=2 15964 requests/s avgLatency=125ms P99Latency=251ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T15:18:43+01:00" level=info msg="Overall for passthrough samples=2 14742 requests/s avgLatency=132ms P99Latency=260ms sum(timeouts)=0 sum(httperrors)=0"

# % ag overall ~/x
# 478:time="2023-09-08T16:12:06+01:00" level=info msg="Overall for http samples=2 30098 requests/s avgLatency=69ms P99Latency=232ms sum(timeouts)=0 sum(httperrors)=0"
# 953:time="2023-09-08T16:22:08+01:00" level=info msg="Overall for edge samples=2 26049 requests/s avgLatency=85ms P99Latency=350ms sum(timeouts)=0 sum(httperrors)=0"
# 1428:time="2023-09-08T16:32:11+01:00" level=info msg="Overall for reencrypt samples=2 27815 requests/s avgLatency=74ms P99Latency=235ms sum(timeouts)=0 sum(httperrors)=0"
# 1903:time="2023-09-08T16:42:14+01:00" level=info msg="Overall for passthrough samples=2 23475 requests/s avgLatency=85ms P99Latency=292ms sum(timeouts)=0 sum(httperrors)=0"

print "* Latency\n";
{
    my @http_latency = (
	{'4.13.8'     => 71 },
	{'4.13.8 haproxy-2.6'     => 69 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 127 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 128 },
	);

    my @edge_latency = (
	{'4.13.8'     => 86 },
	{'4.13.8 haproxy-2.6'     => 85 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 125 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 126 },
	);

    my @reencrypt_latency = (
	{'4.13.8'     =>  70 },
	{'4.13.8 haproxy-2.6'     => 74 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        =>  124 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        =>  125 },
	);

    my @passthrough_latency = (
	{'4.13.8'     =>  87 },
	{'4.13.8 haproxy-2.6'     => 85 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.2'        => 133 },
	{'4.14.0-0.nightly-2023-08-11-055332 haproxy-2.6'        => 132 },
	);

    generate_table("http", "latency (ms)", \@http_latency);
    print "\n";
    generate_table("edge", "latency (ms)", \@edge_latency);
    print "\n";
    generate_table("reencrypt", "latency (ms)", \@reencrypt_latency);
    print "\n";
    generate_table("passthrough", "latency (ms)", \@passthrough_latency);
    print "\n";
}

print q!* Cluster Sizing
#+BEGIN_SRC text
"masterNodesType": "",
"workerNodesType": "",
"masterNodesCount": 3,
"infraNodesType": "",
"workerNodesCount": 3,
"infraNodesCount": 2,
"otherNodesCount": 0,
"totalNodes": 8,
"sdnType": "OVNKubernetes",
#+END_SRC
!;

print "\n";

