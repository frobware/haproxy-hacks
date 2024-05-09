#!/usr/bin/env perl

$| = 1;

use strict;
use warnings;

my $backend_name = shift @ARGV;
my $backend_named_matched = 0;

while (my $line = <>) {
    if ($line =~ /^backend\s+/i) {
        if ($line =~ /$backend_name/i) {
            $backend_named_matched = 1;
        } else {
            last if $backend_named_matched;
        }
    }
    print "$line" if $backend_named_matched;
}
