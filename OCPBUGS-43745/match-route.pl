#!/usr/bin/env perl

use strict;
use warnings;

my $route_name = shift @ARGV or die "Usage: $0 <route-name>\n";

my $in_block = 0;
my $start_pattern = qr/^backend.*\Q$route_name\E/;

while (<>) {
    chomp;

    # Start processing when we match the specific backend.
    if (/$start_pattern/) {
        $in_block = 1;
    }

    # Stop processing when we reach a new backend or non-indented line
    # at column 1
    if ($in_block && /^[^\s]/ && !/$start_pattern/) {
        $in_block = 0;
    }

    print "$_\n" if $in_block && /\S/;
}
