#!/usr/bin/perl

use strict;

my $v;

while (<>) {
    chomp;
    my ($a, $b) = split;
    $v = $a unless defined $v;
    my $harumpf = "";
    if ($a - $v < 0) {
	$harumpf = "**** $a < $v ****";
    }
    print "$a $b (", scalar localtime($b), ") $harumpf\n";
    $v = $a;
}
