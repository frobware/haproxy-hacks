#!/usr/bin/env perl

use strict;
use warnings;
use Term::ANSIColor;

sub run_test {
    my ($enable_keep_alive, $cache_control) = @_;
    print "===================================================\n";
    print "Testing with: ENABLE_KEEPALIVE=$enable_keep_alive USE_CACHE_CONTROL=$cache_control\n";
    print "===================================================\n";

    my $cmd = "ENABLE_KEEPALIVE=$enable_keep_alive USE_CACHE_CONTROL=$cache_control go test -v 2>&1";
    my $output = `$cmd`;
    my $exit_code = $? >> 8;

    print $output;
    print "\nTest completed with exit code: $exit_code\n\n";
    return ($exit_code, $output);
}

my @configs = (
    {enable_keep_alive => "false", cache_control => "false"},
    {enable_keep_alive => "false", cache_control => "true"},
    {enable_keep_alive => "true",  cache_control => "false"},
    {enable_keep_alive => "true",  cache_control => "true"}
);

my %results;
print "Running tests with explicit configurations...\n";

for my $config (@configs) {
    my ($exit_code, $output) = run_test($config->{enable_keep_alive}, $config->{cache_control});
    my $key = "$config->{enable_keep_alive}_$config->{cache_control}";
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
    my $key = "$config->{enable_keep_alive}_$config->{cache_control}";
    my $result = $results{$key};
    my $result_text = $result->{exit_code} == 0 ? colored("PASS", "green") : colored("FAIL", "red");

    # Adjust reuse status based on ENABLE_KEEPALIVE logic
    my $reuse_status = $config->{enable_keep_alive} eq 'true' ? "Enabled" : "Disabled";
    my $cache_status = $config->{cache_control} eq 'true' ? "Yes" : "No";

    printf "%-25s %-25s | %-9d | %-6s\n",
        "$reuse_status",
        $cache_status,
        $result->{exit_code},
        $result_text;
}
