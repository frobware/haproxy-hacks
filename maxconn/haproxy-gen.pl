#!/usr/bin/env perl

use strict;

print "
global
  maxconn 1000
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

for my $i (1..1) {
    print "
frontend alive-${i}
  bind unix@/var/lib/haproxy/run/alive-${i}.sock
\n";

}

for my $i (0..1000) {
    print "
backend backend-${i}
    server server_2 127.0.0.1:80
\n";

}

