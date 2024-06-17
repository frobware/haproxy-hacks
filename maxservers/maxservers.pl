#!/usr/bin/env perl

use strict;
use warnings;
use Data::Dumper;
use Getopt::Long;

$| = 1;

my $haproxy = "ocp-haproxy-2.8.5";
my $config_file = "/tmp/haproxy.config";
my $stats_socket = '/tmp/haproxy.sock';
my $use_server_template = 1; # Default to using server-template

GetOptions(
    'use-server-template!' => \$use_server_template,
) or die "Error in command line arguments\n";

sub gen_config {
    my ($maxconn, $nbthread, $backends, $balance_algorithm, $weight, $nservers, $use_server_template, $fh) = @_;
    $fh->autoflush(1);

    print $fh "
global
  daemon
  stats socket $stats_socket mode 600 level admin
  stats timeout 2m
";

    print $fh "  maxconn $maxconn\n" if $maxconn ne 'auto' && $maxconn > 0;
    print $fh "  nbthread $nbthread\n" if $nbthread > 0;

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

    for my $i (1..$backends) {
        print $fh "
backend backend-$i
  balance $balance_algorithm";

        if ($use_server_template) {
            print $fh "
  server-template _dynamic-pod- 1-$nservers 172.4.0.4:8765 check disabled
" if $nservers > 0;
        } else {
            for my $j (1..$nservers) {
                my $ip = sprintf("172.4.%d.%d", int(($j - 1) / 255), ($j - 1) % 255 + 1);
                print $fh "
  server server-$j $ip:8765 check disabled";
            }
            print $fh "\n";
        }
    }

    print $fh "\n";
    $fh->flush();
}

sub run_command_and_wait {
    my ($cmd) = @_;
    my $pid = fork();

    if ($pid == 0) {
        exec($cmd);
        exit(1);
    } elsif ($pid > 0) {
        waitpid($pid, 0);
        if ($? == -1) {
            print "Failed to execute: $!\n";
            return -1;
        } elsif ($? & 127) {
            printf "Child died with signal %d, %s coredump\n",
                ($? & 127),  ($? & 128) ? 'with' : 'without';
            return -1;
        } else {
            return $? >> 8;
        }
    } else {
        die "Failed to fork: $!";
    }
}

sub haproxy_config_loaded {
    my $cmd = 'show info';
    my $output = `echo "$cmd" | socat $stats_socket stdio 2>/dev/null`;
    if ($output =~ /Pid:\s*(\d+)/) {
        return $1;
    } else {
        return 0;
    }
}

sub kill_haproxy {
    my $haproxy_pid = haproxy_config_loaded();
    if ($haproxy_pid) {
        run_command_and_wait("pkill -TERM -P $haproxy_pid");
        run_command_and_wait("pkill -TERM $haproxy_pid");
    }
    run_command_and_wait("pkill -f $haproxy");
}

sub launch_haproxy {
    unlink $stats_socket if -e $stats_socket;

    my $filename = shift;
    my $haproxy_cmd = "$haproxy -f $filename";
    my $exit_status = run_command_and_wait($haproxy_cmd);
    if ($exit_status != 0) {
        die "HAProxy failed to start. Exit status: $exit_status\n";
    }

    # Wait for HAProxy to be ready.
    for (my $i = 0; $i < 60; $i++) {
        if (-e $stats_socket) {
            last;
        }
        sleep 1;
    }

    die "HAProxy management socket not available" unless -e $stats_socket;

    for (my $i = 0; $i < 60; $i++) {
        last if haproxy_config_loaded();
        sleep 1;
    }

    die "HAProxy failed to load configuration" unless haproxy_config_loaded();

    my $haproxy_process_pid = haproxy_config_loaded();
    my $consecutive_same_memory = 0;
    my $previous_memory = -1;
    my $total_attempts = 0;
    my $max_retries = 5;        # Maximum retries of 5 attempts each
    my $attempt_limit = 5;      # 5 consecutive checks per retry

    # Monitor HAProxy memory usage with strict conditions.
    while ($total_attempts < $max_retries * $attempt_limit) {
        my $current_memory = -1;
        my @output = `pmap -x -p $haproxy_process_pid`;

        foreach my $line (@output) {
            if ($line =~ /total kB\s+\d+\s+(\d+)\s+\d+/) {
                $current_memory = $1;
                last;
            }
        }

        # Update the consecutive same memory count or reset if changed.
        if (defined $previous_memory && $current_memory == $previous_memory) {
            $consecutive_same_memory++;
        } else {
            # Reset to 1 because we have 1 occurrence of this new memory amount.
            $consecutive_same_memory = 1;
        }

        # Update previous memory.
        $previous_memory = $current_memory;

        # Increment the total attempts after each check.
        $total_attempts++;
        print STDERR "Attempt $total_attempts: $consecutive_same_memory consecutive readings of $current_memory kB\n";

        # Check if consecutive readings meet the required count to conclude stability.
        if ($consecutive_same_memory >= 5) {
            print STDERR "Memory usage is stable.\n";
            # Exit the while loop early if memory is stable.
            last;
        }

        # Sleep between pmap attempts to allow for possible changes in memory usage.
        # sleep(int($total_attempts / 5) + 2);

        # Reset and continue if a set of 5 attempts completes without 5 consecutive same readings.
        if ($total_attempts % 5 == 0 && $consecutive_same_memory < 5) {
            print STDERR "Retrying for another set of 5 attempts.\n";
            # Explicit reset here is not necessary but is kept for clarity.
            $consecutive_same_memory = 0;
        }
    }

    # Handle cases where memory did not stabilise after maximum retries.
    die "Memory usage did not stabilise after $max_retries sets of attempts.\n" if $consecutive_same_memory < 5;

    my $rss_mb = int($previous_memory / 1024);
    return ($previous_memory, $rss_mb);
}

my @weights = qw(1);
my @balance_algorithms = qw(leastconn random roundrobin);
my @backends = qw(100 1000 10000);
my @nservers = qw(0 1 2 5 10 100 200 300);
my @nbthreads = qw(4 64);
my @max_connections = qw(50000);

for my $weight (@weights) {
    for my $balance_algorithm (@balance_algorithms) {
        for my $backends (@backends) {
            for my $nservers (@nservers) {
                for my $nbthread (@nbthreads) {
                    for my $maxconn (@max_connections) {
                        print STDERR "Testing: weight=$weight, algorithm=$balance_algorithm, backends=$backends, servers=$nservers, threads=$nbthread, maxconn=$maxconn, use_server_template=$use_server_template\n";
                        kill_haproxy();
                        open(my $fh, '>', $config_file) or die "Could not open file $config_file: $!";
                        gen_config($maxconn, $nbthread, $backends, $balance_algorithm, $weight, $nservers, $use_server_template, $fh);
                        close($fh) or die "Could not close file $config_file: $!";
                        my ($kb, $mb) = launch_haproxy($config_file);
                        print "$weight $balance_algorithm $backends $nservers $nbthread $kb $mb $use_server_template\n";
                        kill_haproxy();
                    }
                }
            }
        }
    }
}
