#!/usr/bin/env perl

use strict;
use Data::Dumper;


sub gen_config {
    my ($maxconn, $nbthread, $backends, $balance_algorithm, $weight, $fh) = @_;

    print $fh "
global
  stats socket /tmp/haproxy.sock mode 600 level admin expose-fd listeners
  stats timeout 2m
";

    undef $maxconn if $maxconn eq "auto";
    print $fh "maxconn $maxconn\n" if $maxconn > 0;
    print $fh "nbthread $nbthread\n" if $nbthread > 0;

    print $fh "
defaults
  timeout connect 5s
  timeout client 30s
  timeout client-fin 1s
  timeout server 300s
  timeout server-fin 1s
  timeout http-request 10s
  timeout http-keep-alive 300s
  timeout tunnel 1h
  option logasap
\n";

    print $fh "
listen stats
  bind :1936
  mode http
  stats enable
  stats refresh 1s
  stats uri /stats
\n";

    for my $i (1..20) {
        print $fh "
frontend alive-${i}
  bind unix@/tmp/haproxy-frontend-${i}.sock
\n";

    }

    for my $i (0..$backends) {
        print $fh "
backend backend-${i}
    balance $balance_algorithm
    server server_0 127.0.0.1:80 weight $weight
    server server_1 127.0.0.1:80 weight $weight
    server server_2 127.0.0.1:80 weight $weight
    server server_3 127.0.0.1:80 weight $weight
    server server_4 127.0.0.1:80 weight $weight
    server server_5 127.0.0.1:80 weight $weight
    server server_6 127.0.0.1:80 weight $weight
    server server_7 127.0.0.1:80 weight $weight
    server server_8 127.0.0.1:80 weight $weight
    server server_9 127.0.0.1:80 weight $weight
\n";
    }

    close $fh or die;
}

sub haproxy_memsize_in_megabytes {
    my $filename = shift;
    my $mb = 0;
    system("pkill haproxy");
    system("sleep 1");
    system("pkill haproxy");
    system("haproxy -D -f $filename");
    system("sleep 1");
    my @output = `pmap -x -p \$(pgrep -n haproxy)`;
    for my $line (@output) {
        next unless $line =~ m/total kB/;
        my @fields = split /\s+/, $line;
        return int($fields[4]);
    }
    die "didn't find memory usage info; do you have haproxy on your PATH?";
}

sub haproxy_actual_maxconn_and_maxsock {
    my @output = `echo "show info" | socat /tmp/haproxy.sock stdio | grep -i -e '^maxsock:' -e '^maxconn:'`;
    my $maxconn = 0;
    my $maxsock = 0;
    for my $line (@output) {
        chomp $line;
        if ($line =~ m/^Maxsock: (\d+)/) {
            $maxsock = $1;
        }
        if ($line =~ m/^Maxconn: (\d+)/) {
            $maxconn = $1;
        }
    }
    return $maxconn, $maxsock;
}

my @weights = qw(1 256);
my @balance_algorithms = qw(random leastconn roundrobin);
my @backends = qw(1000 2000 4000 10000);
my @nbthreads = qw(4 64);
my @max_connections = qw(2000 20000 50000 100000 200000 auto);

# my @weights = qw(1);
# my @balance_algorithms = qw(leastconn);
# my @backends = qw(1000);
# my @nbthreads = qw(4);
# my @max_connections = qw(20000 auto);

for my $weight (@weights) {
    for my $balance_algorithm (@balance_algorithms) {
        for my $backends (@backends) {
            for my $nbthread (@nbthreads) {
                for my $maxconn (@max_connections) {
                    my $filename = "/tmp/haproxy.config";
                    open(FH, '>', $filename)
                        or die "Could not open file $filename: $!";
                    gen_config($maxconn, $nbthread, $backends, $balance_algorithm, $weight, \*FH);
                    my $process_size_in_kb = haproxy_memsize_in_megabytes($filename);
                    my $process_size_in_mb = int($process_size_in_kb / 1000);
                    my ($actual_maxconn, $actual_maxsock) = haproxy_actual_maxconn_and_maxsock();
                    print "$balance_algorithm $weight $backends $nbthread $maxconn $actual_maxconn $actual_maxsock $process_size_in_kb $process_size_in_mb\n";
                }
            }
        }
    }
}
