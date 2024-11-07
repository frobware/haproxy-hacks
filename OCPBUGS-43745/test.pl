#!/usr/bin/env perl

use strict;
use warnings;
use Term::ANSIColor;

sub run_test {
    my ($keep_alive, $cache_control) = @_;
    print "===================================================\n";
    print "Testing with: DISABLE_KEEPALIVE=$keep_alive USE_CACHE_CONTROL=$cache_control\n";
    print "===================================================\n";

    my $cmd = "DISABLE_KEEPALIVE=$keep_alive USE_CACHE_CONTROL=$cache_control go test -v 2>&1";
    my $output = `$cmd`;
    my $exit_code = $? >> 8;

    print $output;
    print "\nTest completed with exit code: $exit_code\n\n";
    return ($exit_code, $output);
}

my @configs = (
    {keep_alive => "false", cache_control => "false"},
    {keep_alive => "false", cache_control => "true"},
    {keep_alive => "true",  cache_control => "false"},
    {keep_alive => "true",  cache_control => "true"}
    );

my %results;
print "Running tests with explicit configurations...\n";

for my $config (@configs) {
    my ($exit_code, $output) = run_test($config->{keep_alive}, $config->{cache_control});
    my $key = "$config->{keep_alive}_$config->{cache_control}";
    $results{$key} = {
        exit_code => $exit_code,
        output => $output
    };
}

print "Test results\n";
print "===================================================\n";
printf "%-25s %-25s | %-9s | %-6s\n",
    "Connection Reuse", "Cache-Control: no-cache", "Exit Code", "Result";
print "-" x 70 . "\n";

for my $config (@configs) {
    my $key = "$config->{keep_alive}_$config->{cache_control}";
    my $result = $results{$key};
    my $result_text = $result->{exit_code} == 0 ? colored("PASS", "green") : colored("FAIL", "red");

    my $reuse_status = $config->{keep_alive} eq 'false' ? "Enabled" : "Disabled";
    my $cache_status = $config->{cache_control} eq 'true' ? "Yes" : "No";

    printf "%-25s %-25s | %-9d | %-6s\n",
        "$reuse_status",
        $cache_status,
        $result->{exit_code},
        $result_text;
}
