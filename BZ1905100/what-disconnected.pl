#!/usr/bin/env perl

use strict;
use Data::Dumper;

my $file_a = $ARGV[0];
my $file_b = $ARGV[1];

sub IPv4cmp {
    my @a  = split /\./, $a;
    my @b = split /\./, $b;

    return $a[0] <=> $b[0] ||
           $a[1] <=> $b[1] ||
           $a[2] <=> $b[2] ||
           $a[3] <=> $b[3];
}

sub PeerIPv4cmp {
    my @a  = split /\./, $a->[0];
    my @b = split /\./, $b->[0];

    return $a[0] <=> $b[0] ||
	$a[1] <=> $b[1] ||
	$a[2] <=> $b[2] ||
	$a[3] <=> $b[3] ||
	$a->[1] <=> $b->[1];
}

sub slurpfile {
    my $filename = shift;
    my %established_connections;

    open(FH, '<', $filename) or die $!;

    while (<FH>) {
	chomp;
	next unless /ESTAB/;
	my @words = split /\s+/;
	$established_connections{"$words[0] local:$words[4] peer:$words[5]"} = { local => "$words[4]", remote => "$words[5]" };
    }

    return %established_connections;
}

my %file_a_connections = slurpfile("$file_a");
my %file_b_connections = slurpfile("$file_b");
my @dropped;

for my $k (keys %file_a_connections) {
    unless (exists($file_b_connections{"$k"})) {
	my ($ip_addr, $port) = split(/:/, $file_a_connections{$k}{"remote"});
	push(@dropped, [ $ip_addr, $port ]);
	#print("GONE $ip_addr $port\n");
	#push(@dropped, $file_a_connections{$k}{"remote"});
    }
}

for my $g (sort PeerIPv4cmp @dropped) {
    print "ESTABLISHED connection GONE: $g->[0]:$g->[1]\n";
}

print "$file_a has ", scalar %file_a_connections, " ESTABLISHED connections\n";
print "$file_b has ", scalar %file_b_connections, " ESTABLISHED connections\n";
print scalar @dropped, " ESTABLISHED connections have gone\n";

my %peers;

for my $peer_info (@dropped) {
    my ($ip_addr, $port) = split /:/, $peer_info->[0], $peer_info->[1];
    $peers{$ip_addr} +=1
}

my @ip = sort IPv4cmp keys %peers;

for my $ip (@ip) {
    print "$ip -- $peers{$ip} dropped connections\n";
}
