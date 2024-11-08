#!/usr/bin/env perl

use strict;
use warnings;
use Term::ANSIColor;

if (@ARGV == 0) {
    die "Usage: $0 <delay_in_seconds> ...\nExample: $0 30 35 40 45 50 55 60\n";
}

# Variables to store OpenShift and HAProxy version information.
my ($openshift_version, $haproxy_version);

sub run_test {
    my ($delay) = @_;
    print "=============================================\n";
    print "Testing with: REQUEST_DELAY=${delay}s ENABLE_KEEPALIVE=true\n";
    print "=============================================\n";

    my $cmd = "REQUEST_DELAY=${delay}s ENABLE_KEEPALIVE=true go test -v 2>&1";
    my $output = `$cmd`;
    my $exit_code = $? >> 8;

    # Extract OpenShift and HAProxy version info if not already set.
    my @lines = split "\n", $output;
    if (!$openshift_version && !$haproxy_version && @lines >= 3) {
        $openshift_version = $lines[1];
        $haproxy_version = $lines[2];
    }

    print "\nTest completed with exit code: $exit_code\n\n";
    return ($exit_code, $output);
}

my %results;
print "Running tests with varying REQUEST_DELAY values...\n";

# Iterate over each delay value provided as a CLI argument
foreach my $delay (@ARGV) {
    my ($exit_code, $output) = run_test($delay);
    $results{$delay} = {
        exit_code => $exit_code,
        output => $output
    };
}

print "Test results\n";
print "=============================================\n";
if ($openshift_version && $haproxy_version) {
    print "OpenShift Cluster Version: $openshift_version\n";
    print "HAProxy Version: $haproxy_version\n";
    print "---------------------------------------------\n";
}
printf "%-15s | %-9s | %-6s\n", "REQUEST_DELAY", "Exit Code", "Result";
print "-" x 40 . "\n";

for my $delay (sort { $a <=> $b } keys %results) {
    my $result = $results{$delay};
    my $result_text = $result->{exit_code} == 0 ? colored("PASS", "green") : colored("FAIL", "red");

    printf "%-15s | %-9d | %-6s\n", "${delay}s", $result->{exit_code}, $result_text;
}
