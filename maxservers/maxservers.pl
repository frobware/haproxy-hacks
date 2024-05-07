#!/usr/bin/env perl

$| = 1;

use strict;
use Data::Dumper;

my $haproxy = "ocp-haproxy-2.8.5";

sub gen_config {
    my ($maxconn, $nbthread, $backends, $balance_algorithm, $weight, $nservers, $fh) = @_;

    $fh->autoflush(1);

    print $fh "
global
  daemon
  stats socket /tmp/haproxy.sock mode 600 level admin
  stats timeout 2m
";

    undef $maxconn if $maxconn eq "auto";
    print $fh "  maxconn $maxconn\n" if $maxconn > 0;
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
        print $fh "
    server-template _dynamic-pod- 1-$nservers 172.4.0.4:8765 check disabled
"
            if $nservers > 0;
    }

    print $fh "\n";
    $fh->flush();
}

sub run_command_and_wait {
    my ($cmd) = @_;
    my $pid = fork();

    if ($pid == 0) {
        # This is the child process.
        exec($cmd);
        # If exec fails, exit with a non-zero status.
        exit(1);
    } elsif ($pid > 0) {
        # This is the parent process.
        # Wait for the child process to terminate.
        waitpid($pid, 0);
        # Check the exit status of the child process.
        if ($? == -1) {
            print "Failed to execute: $!\n";
            return -1;
        } elsif ($? & 127) {
            printf "Child died with signal %d, %s coredump\n",
                ($? & 127),  ($? & 128) ? 'with' : 'without';
            return -1;
        } else {
            my $exit_code = $? >> 8;
            return $exit_code;
        }
    } else {
        # Fork failed.
        die "Failed to fork: $!";
    }
}

sub haproxy_config_loaded {
    my $stats_socket = '/tmp/haproxy.sock';
    my $cmd = 'show info';
    my $output = `echo "$cmd" | socat $stats_socket stdio 2>/dev/null`;
    if ($output =~ /Pid:\s*(\d+)/) {
        return $1;              # Return the PID if found
    } else {
        return 0;               # Return false if PID not found
    }
}

sub launch_haproxy {
    if (-e '/tmp/haproxy.sock') {
        unlink '/tmp/haproxy.sock'
            or die "Failed to delete socket: $!";
    }

    my $filename = shift;
    my $haproxy_cmd = "$haproxy -f $filename";
    my $exit_status = run_command_and_wait($haproxy_cmd);
    if ($exit_status != 0) {
        print "HAProxy failed to start. Exit status: $exit_status\n";
        exit($exit_status);
    }

    # Wait for HAProxy to be ready.
    my $socket_ready = 0;
    for (my $i = 0; $i < 60; $i++) {
        if (-e '/tmp/haproxy.sock') {
            $socket_ready = 1;
            last;
        }
        sleep 1;
    }

    die "HAProxy management socket not available" unless $socket_ready;

    # Wait for HAProxy to load the configuration.
    for (my $i = 0; $i < 60; $i++) {
        last if haproxy_config_loaded();
        sleep 1;
    }

    die "HAProxy failed to load configuration" unless haproxy_config_loaded();

    my $haproxy_process_pid = haproxy_config_loaded();
    my $consecutive_same_memory = 0;
    my $previous_memory = -1;

    # This is a bit of paranoia. Let's ensure haproxy isn't growing
    # memory (i.e., has it actually finished parsing?).
    for (my $attempt = 1; $attempt <= 5; $attempt++) {
        my @output = `pmap -x -p $haproxy_process_pid`;
        my $current_memory = -1;
        foreach my $line (@output) {
            if ($line =~ /total kB\s+\d+\s+(\d+)\s+\d+/) {
                my $rss_kb = $1;  # This captures the RSS memory in kB
                $current_memory = $rss_kb;
                if ($rss_kb == $previous_memory) {
                    $consecutive_same_memory++;
                } else {
                    $consecutive_same_memory = 0;
                }
                $previous_memory = $rss_kb;
            }
        }
        last if $consecutive_same_memory == 5;
        sleep 0.5;
    }

    die "Didn't find consistent memory usage info ($consecutive_same_memory)"
        if $consecutive_same_memory < 4;

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
                        my $filename = "/tmp/haproxy.config";
                        # Kill any old process.
                        run_command_and_wait("pkill -f $haproxy");
                        run_command_and_wait("pkill -f $haproxy");
                        open(my $fh, '>', $filename) or die "Could not open file $filename: $!";
                        gen_config($maxconn, $nbthread, $backends, $balance_algorithm, $weight, $nservers, $fh);
                        close($fh) or die "Could not close file $filename: $!";
                        my ($kb, $mb) = launch_haproxy($filename);
                        print "$weight $balance_algorithm $backends $nservers $nbthread $kb $mb\n";
                        # my $pause_for_interative_poking = <>;
                    }
                }
            }
        }
    }
}
