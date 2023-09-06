#!/usr/bin/env perl

use strict;
use warnings;

sub generate_table {
    my ($termination_type, $metric, $values) = @_;

    print "** $termination_type\n";

    print "| Release | $metric | % Diff from 4.13.9 |\n";
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

print "* Requests/s\n";

{
   # 2.6.13

   # 8305:     2  INFO Summary for "http" samples=1 76,156 requests/s avgLatency=57ms P99Latency=80ms sum(timeouts)=1626 sum(httperrors)=0
   # 8308:     5  INFO Summary for "http" samples=2 73,365 requests/s avgLatency=46ms P99Latency=73ms sum(timeouts)=3559 sum(httperrors)=0
   # 8311:     8  INFO Summary for "edge" samples=2 57,320 requests/s avgLatency=60ms P99Latency=90ms sum(timeouts)=0 sum(httperrors)=0
   # 8314:    11  INFO Summary for "reencrypt" samples=2 59,930 requests/s avgLatency=50ms P99Latency=519ms sum(timeouts)=0 sum(httperrors)=0
   # 8317:    14  INFO Summary for "passthrough" samples=2 74,131 requests/s avgLatency=7ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0

   # 2.2.24

   # 8323:     2  INFO Summary for "http" samples=1 76,495 requests/s avgLatency=46ms P99Latency=67ms sum(timeouts)=1579 sum(httperrors)=0
   # 8326:     5  INFO Summary for "http" samples=2 70,722 requests/s avgLatency=46ms P99Latency=66ms sum(timeouts)=3574 sum(httperrors)=0
   # 8329:     8  INFO Summary for "edge" samples=2 58,066 requests/s avgLatency=60ms P99Latency=82ms sum(timeouts)=0 sum(httperrors)=0
   # 8332:    11  INFO Summary for "reencrypt" samples=2 63,849 requests/s avgLatency=35ms P99Latency=55ms sum(timeouts)=0 sum(httperrors)=0
   # 8335:    14  INFO Summary for "passthrough" samples=2 78,322 requests/s avgLatency=6ms P99Latency=8ms sum(timeouts)=0 sum(httperrors)=0

    my @edge_rps = (
	{'4.13.9'     => 70.74},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 57.32},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 58.06},
	);

    my @http_rps = (
	{'4.13.9'     => 117.25},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 73.65},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 70.72},
	);

    my @passthrough_rps = (
	{'4.13.9'     => 201.01 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 74.13 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 78.32},
	);

    my @reencrypt_rps = (
	{'4.13.9'     => 72.51 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 59.93 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 63.84},
	);

    generate_table("edge", "request/s (x1000)", \@edge_rps);
    print "\n";
    generate_table("http", "request/s (x1000)", \@http_rps);
    print "\n";
    generate_table("passthrough", "request/s (x1000)", \@passthrough_rps);
    print "\n";
    generate_table("reencrypt", "reencrypt/s (x1000)", \@reencrypt_rps);
    print "\n";
}

print "* Latency\n";

   # 2.6.13

   # 8305:     2  INFO Summary for "http" samples=1 76,156 requests/s avgLatency=57ms P99Latency=80ms sum(timeouts)=1626 sum(httperrors)=0
   # 8308:     5  INFO Summary for "http" samples=2 73,365 requests/s avgLatency=46ms P99Latency=73ms sum(timeouts)=3559 sum(httperrors)=0
   # 8311:     8  INFO Summary for "edge" samples=2 57,320 requests/s avgLatency=60ms P99Latency=90ms sum(timeouts)=0 sum(httperrors)=0
   # 8314:    11  INFO Summary for "reencrypt" samples=2 59,930 requests/s avgLatency=50ms P99Latency=519ms sum(timeouts)=0 sum(httperrors)=0
   # 8317:    14  INFO Summary for "passthrough" samples=2 74,131 requests/s avgLatency=7ms P99Latency=13ms sum(timeouts)=0 sum(httperrors)=0

   # 2.2.24

   # 8323:     2  INFO Summary for "http" samples=1 76,495 requests/s avgLatency=46ms P99Latency=67ms sum(timeouts)=1579 sum(httperrors)=0
   # 8326:     5  INFO Summary for "http" samples=2 70,722 requests/s avgLatency=46ms P99Latency=66ms sum(timeouts)=3574 sum(httperrors)=0
   # 8329:     8  INFO Summary for "edge" samples=2 58,066 requests/s avgLatency=60ms P99Latency=82ms sum(timeouts)=0 sum(httperrors)=0
   # 8332:    11  INFO Summary for "reencrypt" samples=2 63,849 requests/s avgLatency=35ms P99Latency=55ms sum(timeouts)=0 sum(httperrors)=0
   # 8335:    14  INFO Summary for "passthrough" samples=2 78,322 requests/s avgLatency=6ms P99Latency=8ms sum(timeouts)=0 sum(httperrors)=0
{
    my @edge_latency = (
	{'4.13.9'     => 50.96},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 60},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 60},
	);

    my @http_latency = (
	{'4.13.9'     => 32.17},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 46},
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 46},
	);

    my @passthrough_latency = (
	{'4.13.9'     => 18.06 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 7 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 6 },
	);

    my @reencrypt_latency = (
	{'4.13.9'     =>  50.19 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.6'        => 50 },
	{'4.14.0-0.nightly-2023-09-02-132842 haproxy-2.2'        => 34 },
	);

    generate_table("edge", "latency (ms)", \@edge_latency);
    print "\n";
    generate_table("http", "latency (ms)", \@http_latency);
    print "\n";
    generate_table("passthrough", "latency (ms)", \@passthrough_latency);
    print "\n";
    generate_table("reencrypt", "latency (ms)", \@reencrypt_latency);
    print "\n";
}
