#!/usr/bin/env perl

use File::Grep;
use Data::Dumper qw(Dumper);

my %nodes;

while (<STDIN>) {
    next unless /openshift/;

    if (/([^:]+)/) {
	my $base = "/Users/amcdermo/net/02867376/namespaces";
	my $file = "$1";

	@r = `rg 'restartCount:.\([1-9]+\)' -r '\$1' $file`;
	@f = `rg 'nodeName:.\(ip-[^ ]+\)' -r '\$1' $file`;

	for my $i (0 .. $#r) {
	    print;

	    chomp $f[$i];
	    chomp $r[$i];

	    next unless  length($f[$i]) > 0;
	    $f[$i] =~ s/^\s+|\s+$//g;
	    $r[$i] =~ s/^\s+|\s+$//g;

	    printf "\trestartCount: $r[$i] nodeName: $f[$i]\n";

	    $nodes{$f[$i]} += $r[i];

	print "\n";
	}
    }
}

print Dumper("Aggregated restarts", \%nodes), "\n";
