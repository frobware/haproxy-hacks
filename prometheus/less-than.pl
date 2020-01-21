#!/usr/bin/perl

use strict;

my $v;

while (<>) {
    chomp;
    my ($a, $b) = split;
    $v = $a unless defined $v;
    #print "$a >= $v\n";
    if ($a - $v < 0) {
	die "error: $a < $v";
    }
    $v = $a;
}
