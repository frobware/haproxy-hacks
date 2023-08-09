#!/usr/bin/env perl

use strict;
use warnings;

my $NTH = shift;  # Get the NTH value from command line argument
die "Usage: $0 NTH" unless defined $NTH;

my $cert_number = 0;
my $cert_block = '';

while (<STDIN>) {
    if (/-----BEGIN CERTIFICATE-----/) {
        $cert_number++;
        if ($cert_number == $NTH) {
            $cert_block .= $_;
        }
    } elsif (/-----END CERTIFICATE-----/) {
        if ($cert_number == $NTH) {
            $cert_block .= $_;
            print $cert_block;
            last;
        }
    } elsif ($cert_number == $NTH) {
        $cert_block .= $_;
    }
}
