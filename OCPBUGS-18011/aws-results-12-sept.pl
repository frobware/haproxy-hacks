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

print "* Requests/s\n";
{
    my @http_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 167000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 168000},
	);

    my @edge_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 116000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 113000},
	);

    my @reencrypt_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 100000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 99500},
	);

    my @passthrough_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 294000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 292000},
	);

    generate_table("http", "request/s", \@http_rps);
    print "\n";
    generate_table("edge", "request/s", \@edge_rps);
    print "\n";
    generate_table("reencrypt", "reencrypt/s", \@reencrypt_rps);
    print "\n";
    generate_table("passthrough", "request/s", \@passthrough_rps);
    print "\n";
}

print "* Latency\n";
{
    my @http_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 25.6},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 29.2},
	);

    my @edge_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 38.4},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 46.1},
	);

    my @reencrypt_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 46.7},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 45.1},
	);

    my @passthrough_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 18.9},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 15.7},
	);

    generate_table("http", "P95 latency (ms)", \@http_latency);
    print "\n";
    generate_table("edge", "P95 latency (ms)", \@edge_latency);
    print "\n";
    generate_table("reencrypt", "P95 latency (ms)", \@reencrypt_latency);
    print "\n";
    generate_table("passthrough", "P95 latency (ms)", \@passthrough_latency);
    print "\n";
}

print "\n";

