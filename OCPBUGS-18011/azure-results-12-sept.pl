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
	{'4.13.0-0.nightly-2023-09-05-135358'     => 150000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 144000},
	);

    my @edge_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 114000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 106000},
	);

    my @reencrypt_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 103000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 96300},
	);

    my @passthrough_rps = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 188000},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 177000},
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
	{'4.13.0-0.nightly-2023-09-05-135358'     => 8.28},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 8.25},
	);

    my @edge_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 10.7},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 11.5},
	);

    my @reencrypt_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 11.7},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 15.1},
	);

    my @passthrough_latency = (
	{'4.13.0-0.nightly-2023-09-05-135358'     => 5.7},
	{'4.14.0-0.nightly-2023-09-02 132842'        => 5.6},
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

