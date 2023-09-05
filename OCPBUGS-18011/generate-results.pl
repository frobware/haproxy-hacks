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
    # data from row 201: (reverse '(74.8 73.72 75.06 72.07 76.84 70.74))
    my @edge_rps = (
	{'4.13.9'     => 70.74},
	{'ec1'        => 76.84},
	{'ec2-2.2.24' => 72.07},
	{'ec2-2.6.13' => 75.06},
	{'0828-2.2.24'=> 73.72},
	{'0828-2.6.13'=> 74.08},
	);

    # data from row 202: (reverse '(103.83 105.33 113.51 117.88 122.3 117.25))
    my @http_rps = (
	{'4.13.9'     => 117.25},
	{'ec1'        => 122.30},
	{'ec2-2.2.24' => 117.88},
	{'ec2-2.6.13' => 113.51},
	{'0828-2.2.24'=> 105.33},
	{'0828-2.6.13'=> 103.83},
	);

    # data from row 203: (reverse '(184.74 188.75 197.61 210.72 226.25 201.01))
    my @passthrough_rps = (
	{'4.13.9'     => 201.01 },
	{'ec1'        => 226.25 },
	{'ec2-2.2.24' => 210.71 },
	{'ec2-2.6.13' => 197.61 },
	{'0828-2.2.24'=> 188.75 },
	{'0828-2.6.13'=> 184.74 },
	);

    # data from row 204: (reverse '(72.19 70.73 71.44 68.69 73.74 72.51))
    my @reencrypt_rps = (
	{'4.13.9'     => 72.51 },
	{'ec1'        => 73.74 },
	{'ec2-2.2.24' => 68.69 },
	{'ec2-2.6.13' => 71.44 },
	{'0828-2.2.24'=> 70.73 },
	{'0828-2.6.13'=> 72.19 },
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

{
    # data from row 201: (reverse '(56.95 53.22 49.02 53.68 48.32 50.96))
    my @edge_latency = (
	{'4.13.9'     => 50.96},
	{'ec1'        => 48.32},
	{'ec2-2.2.24' => 53.68},
	{'ec2-2.6.13' => 49.02},
	{'0828-2.2.24'=> 53.22},
	{'0828-2.6.13'=> 56.95},
	);

    # data from row 202: (reverse '(45.39 36.32 32.77 41.73 31.94 32.17))
    my @http_latency = (
	{'4.13.9'     => 32.17},
	{'ec1'        => 31.94},
	{'ec2-2.2.24' => 41.73},
	{'ec2-2.6.13' => 32.77},
	{'0828-2.2.24'=> 36.32},
	{'0828-2.6.13'=> 45.39},
	);

    # data from row 203: (reverse '(20.3 21.01 18.9 17.36 15.86 18.06))
    my @passthrough_latency = (
	{'4.13.9'     => 18.06 },
	{'ec1'        => 15.86 },
	{'ec2-2.2.24' => 17.36 },
	{'ec2-2.6.13' => 18.90 },
	{'0828-2.2.24'=> 21.01 },
	{'0828-2.6.13'=> 20.30 },
	);

    # data from row 204: (reverse '(97.52 58.24 108.11 52.67 54.5 50.19))
    my @reencrypt_latency = (
	{'4.13.9'     =>  50.19 },
	{'ec1'        =>  54.50 },
	{'ec2-2.2.24' =>  52.67 },
	{'ec2-2.6.13' =>  108.11 },
	{'0828-2.2.24'=>  58.24 },
	{'0828-2.6.13'=>  97.52 },
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
