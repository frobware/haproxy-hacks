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

  # "masterNodesType": "m5.8xlarge",
  # "workerNodesType": "m5.2xlarge",
  # "masterNodesCount": 3,
  # "infraNodesType": "r5.2xlarge",
  # "workerNodesCount": 24,
  # "infraNodesCount": 3,
  # "otherNodesCount": 0,
  # "totalNodes": 30,
  # "sdnType": "OVNKubernetes",

# % grep Overall testrun-4.13.8-haproxy-2.2.txt
# time="2023-09-07T14:32:02+01:00" level=info msg="Overall for http samples=2 82233 requests/s avgLatency=43ms P99Latency=107ms sum(timeouts)=3657 sum(httperrors)=0"
# time="2023-09-07T14:42:16+01:00" level=info msg="Overall for edge samples=2 57876 requests/s avgLatency=60ms P99Latency=88ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T14:44:29+01:00" level=info msg="Overall for reencrypt samples=2 60956 requests/s avgLatency=54ms P99Latency=92ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T14:54:43+01:00" level=info msg="Overall for passthrough samples=2 76726 requests/s avgLatency=7ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.13.8-haproxy-2.6.txt
# time="2023-09-07T15:28:20+01:00" level=info msg="Overall for http samples=2 82819 requests/s avgLatency=44ms P99Latency=144ms sum(timeouts)=3670 sum(httperrors)=0"
# time="2023-09-07T15:38:34+01:00" level=info msg="Overall for edge samples=2 57860 requests/s avgLatency=62ms P99Latency=108ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:40:47+01:00" level=info msg="Overall for reencrypt samples=2 57007 requests/s avgLatency=84ms P99Latency=1036ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:51:01+01:00" level=info msg="Overall for passthrough samples=2 105219 requests/s avgLatency=4ms P99Latency=9ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.14-haproxy-2.6.txt
# time="2023-09-07T15:09:00+01:00" level=info msg="Overall for http samples=2 71854 requests/s avgLatency=50ms P99Latency=69ms sum(timeouts)=3668 sum(httperrors)=0"
# time="2023-09-07T15:19:15+01:00" level=info msg="Overall for edge samples=2 58312 requests/s avgLatency=63ms P99Latency=84ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:21:28+01:00" level=info msg="Overall for reencrypt samples=2 58751 requests/s avgLatency=75ms P99Latency=847ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:31:42+01:00" level=info msg="Overall for passthrough samples=2 76527 requests/s avgLatency=4ms P99Latency=9ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.14-haproxy-2.2.txt
# time="2023-09-07T16:11:03+01:00" level=info msg="Overall for http samples=2 72538 requests/s avgLatency=47ms P99Latency=74ms sum(timeouts)=3719 sum(httperrors)=0"
# time="2023-09-07T16:21:17+01:00" level=info msg="Overall for edge samples=2 57935 requests/s avgLatency=62ms P99Latency=92ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T16:23:30+01:00" level=info msg="Overall for reencrypt samples=2 59777 requests/s avgLatency=49ms P99Latency=84ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T16:33:44+01:00" level=info msg="Overall for passthrough samples=2 75127 requests/s avgLatency=8ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0"

print "* Requests/s\n";
{
    my @http_rps = (
	{'4.13.8'     => 82233},
	{'4.13.8 haproxy-2.6'     => 82819},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 71854},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 72538},
	);

    my @edge_rps = (
	{'4.13.8'     => 57876},
	{'4.13.8 haproxy-2.6'     => 57860},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 58312},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 57935},
	);

    my @reencrypt_rps = (
	{'4.13.8'     => 60956 },
	{'4.13.8 haproxy-2.6'     => 57007},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 58751 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 59777},
	);

    my @passthrough_rps = (
	{'4.13.8'     =>  76726 },
	{'4.13.8 haproxy-2.6'     => 105219},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 76527 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 75127},
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

# % grep Overall testrun-4.13.8-haproxy-2.2.txt
# time="2023-09-07T14:32:02+01:00" level=info msg="Overall for http samples=2 82233 requests/s avgLatency=43ms P99Latency=107ms sum(timeouts)=3657 sum(httperrors)=0"
# time="2023-09-07T14:42:16+01:00" level=info msg="Overall for edge samples=2 57876 requests/s avgLatency=60ms P99Latency=88ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T14:44:29+01:00" level=info msg="Overall for reencrypt samples=2 60956 requests/s avgLatency=54ms P99Latency=92ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T14:54:43+01:00" level=info msg="Overall for passthrough samples=2 76726 requests/s avgLatency=7ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.13.8-haproxy-2.6.txt
# time="2023-09-07T15:28:20+01:00" level=info msg="Overall for http samples=2 82819 requests/s avgLatency=44ms P99Latency=144ms sum(timeouts)=3670 sum(httperrors)=0"
# time="2023-09-07T15:38:34+01:00" level=info msg="Overall for edge samples=2 57860 requests/s avgLatency=62ms P99Latency=108ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:40:47+01:00" level=info msg="Overall for reencrypt samples=2 57007 requests/s avgLatency=84ms P99Latency=1036ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:51:01+01:00" level=info msg="Overall for passthrough samples=2 105219 requests/s avgLatency=4ms P99Latency=9ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.14-haproxy-2.6.txt
# time="2023-09-07T15:09:00+01:00" level=info msg="Overall for http samples=2 71854 requests/s avgLatency=50ms P99Latency=69ms sum(timeouts)=3668 sum(httperrors)=0"
# time="2023-09-07T15:19:15+01:00" level=info msg="Overall for edge samples=2 58312 requests/s avgLatency=63ms P99Latency=84ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:21:28+01:00" level=info msg="Overall for reencrypt samples=2 58751 requests/s avgLatency=75ms P99Latency=847ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T15:31:42+01:00" level=info msg="Overall for passthrough samples=2 76527 requests/s avgLatency=4ms P99Latency=9ms sum(timeouts)=0 sum(httperrors)=0"

# % grep Overall testrun-4.14-haproxy-2.2.txt
# time="2023-09-07T16:11:03+01:00" level=info msg="Overall for http samples=2 72538 requests/s avgLatency=47ms P99Latency=74ms sum(timeouts)=3719 sum(httperrors)=0"
# time="2023-09-07T16:21:17+01:00" level=info msg="Overall for edge samples=2 57935 requests/s avgLatency=62ms P99Latency=92ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T16:23:30+01:00" level=info msg="Overall for reencrypt samples=2 59777 requests/s avgLatency=49ms P99Latency=84ms sum(timeouts)=0 sum(httperrors)=0"
# time="2023-09-07T16:33:44+01:00" level=info msg="Overall for passthrough samples=2 75127 requests/s avgLatency=8ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0"

print "* Latency\n";
{
    my @http_latency = (
	{'4.13.8'     => 43},
	{'4.13.8 haproxy 2.6'     => 44},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 50 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 47 },
	);

    my @edge_latency = (
	{'4.13.8'     => 60},
	{'4.13.8 haproxy-2.6'     => 62},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 63 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 62 },
	);

    my @reencrypt_latency = (
	{'4.13.8'     =>  54 },
	{'4.13.8 haproxy-2.6'     => 84},
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        =>  75 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        =>  49 },
	);

    my @passthrough_latency = (
	{'4.13.8'     => 7 },
	{'4.13.8 haproxy-2.6'     => 4 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.6'        => 4 },
	{'4.14.0-0.nightly-2023-09-02 haproxy-2.2'        => 8 },
	);

    generate_table("edge", "latency (ms)", \@edge_latency);
    print "\n";
    generate_table("http", "latency (ms)", \@http_latency);
    print "\n";
    generate_table("reencrypt", "latency (ms)", \@reencrypt_latency);
    print "\n";
    generate_table("passthrough", "latency (ms)", \@passthrough_latency);
    print "\n";
}

print q!* Cluster Sizing
#+BEGIN_SRC text
"masterNodesType": "m5.8xlarge",
"workerNodesType": "m5.2xlarge",
"masterNodesCount": 3,
"infraNodesType": "r5.2xlarge",
"workerNodesCount": 24,
"infraNodesCount": 3,
"otherNodesCount": 0,
"totalNodes": 30,
"sdnType": "OVNKubernetes",
#+END_SRC
!;

print "\n";

