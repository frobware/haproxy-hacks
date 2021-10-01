#!/usr/bin/env perl

use strict;
use Data::Dumper;
use File::stat;

sub trim_quoted {
    return substr($_[0], 1, length($_[0])-2);
}

my $reload_marker_start = "calling reload function";
my $reload_marker_end = $reload_marker_start;

my $synchronized = 0;
my @reloads;
my @f;				# nested input

sub rationalise {
    my ($fh, $line, $current_reload) = @_;
    return if $line =~ /committing router certificate manager changes/;
    return if $line =~ /router certificate manager config committed/;
    return if $line =~ /probehttp.go/;
    return if $line =~ /health check ok/i;

    $fh->print($line) or die $!;

    push(@{$current_reload->{input}}, $line);

    if ($line =~ /MODIFIED/) {
	my @words = split(/\s+/, $line);
	my ($x, $name) = split(/=/, $words[9]);
	$name = trim_quoted($name);
	my ($x, $namespace) = split(/=/, $words[10]);
	$namespace = trim_quoted($namespace);
	my $route = "$name.$namespace";
	$current_reload->{routes}->{$route}->{reason} = $line;
    } elsif ($line =~/adding route/) {
	die;			# unhandled and not seen in the
				# current data set.
    }
}

while ($_ = scalar @f ? shift @f : <STDIN>) {
    if (not $synchronized && /router state synchronized for the first time/) {
	$synchronized = 1;
    }

    next unless $synchronized;

    if (/${reload_marker_start}/) {
	my @words = split(/\s+/);
	my $timestamp = $words[1];
	$timestamp =~ s/\..*//g;
	$timestamp =~ s/://g;
	my $filename = "$words[0]-${timestamp}.rr";
	$filename = substr($filename, 1, length($filename));
	open(OF, '>', "$filename") or die $!;
	# print OF $_ or die $!;
	my $current_reload_time = $timestamp;
	my $reload_details = { filename => $filename, timestamp => $timestamp, routes => {}, input => [] };
	push(@reloads, $reload_details);
	while (<STDIN>) {
	    if (/${reload_marker_end}/) {
		close(OF) or die $!;
		unshift(@f, $_);
		last;
	    } else {
		rationalise(\*OF, "$_", $reload_details);
	    }
	}
    }
}

my $last_timestamp;

for my $reload (@reloads) {
    my $this_timestamp = $reload->{timestamp};
    if ($last_timestamp == 0) {
	$last_timestamp = $reload->{timestamp};
    }

    my $delta = $this_timestamp - $last_timestamp;
    printf("RELOAD %d +%ds -- file %s\n", $this_timestamp, $delta, $reload->{filename});

    my @sorted_routes = sort { $reload->{routes}->{$a} <=> $reload->{routes}->{$b} } keys %{$reload->{routes}};

    for my $key (@sorted_routes) {
	my $reason = $reload->{routes}->{$key}->{reason};

	if (defined $reason) {
	    chomp $reason;
	    printf("\t%s\n\t\tREASON:<<<%s>>>\n", $key, $reason)
	} else {
	    printf("\t%s\n\t\tREASON:<<<%s>>>\n", $key, "***UNKNOWN***");
	}
    }

    $last_timestamp = $this_timestamp;

    print("\n");
}
