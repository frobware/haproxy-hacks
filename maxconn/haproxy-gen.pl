#!/usr/bin/env perl

use strict;

my $maxconn = $ENV{"MAXCONN"} || 1000;
my $nbthread = $ENV{"NBTHREAD"} || 64;
my $backends = $ENV{"BACKENDS"} || 1000;

print "
global
  maxconn $maxconn
  nbthread $nbthread
  stats socket /tmp/haproxy.sock mode 600 level admin expose-fd listeners
  stats timeout 2m

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

print "
listen stats
  bind :1936
  log global
  option httplog
  mode http
  stats enable
  stats refresh 1s
  stats uri /stats
\n";

for my $i (1..20) {
    print "
frontend alive-${i}
  bind unix@/tmp/haproxy-frontend-${i}.sock
\n";

}

for my $i (0..$backends) {
    print "
backend backend-${i}
    balance random
    server server_2 127.0.0.1:80 weight 256
\n";

}

