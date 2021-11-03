#!/usr/bin/env perl

# Usage:

# Generate an OpenShift-esque haproxy.config
#
# $ ./generate-haproxy-config.pl --balance-algorithm=random --proxies=10 --servers=1 --weight=256 --output-dir=/tmp/random --crt-file ~/domain.pem

# Validate config
#
# $ haproxy -c -f /tmp/random/conf/haproxy.config
# Configuration file is valid

use strict;
use warnings;
use Getopt::Long;
use File::Basename;
use Digest::MD5 qw(md5_hex);

my $proxies = 1;
my $balance_algorithm = "random";
my $crt_file;
my $namespace = "default";
my $output_dir;
my $public_http_port = 18080;
my $public_https_port = 18443;

my $fe_sni_port = 10444;
my $be_no_sni_port = 10443;

my $servers = 1;
my $route_type = "edge";
my $target_host = "127.0.0.1";
my $target_port = 9001;
my $weight = 256;

GetOptions("proxies=i" => \$proxies,
	   "balance-algorithm=s", \$balance_algorithm,
	   "crt-file=s", \$crt_file,
	   "http-port=i" => \$public_http_port,
	   "https-port=i" => \$public_https_port,
	   "namespace=s", \$namespace,
	   "output-dir=s" => \$output_dir,
	   "servers=i" => \$servers,
	   "target-host=s" => \$target_host,
	   "target-port=i" => \$target_port,
	   "weight=i" => \$weight)
    or die("error parsing arguments\n");

sub write_to_file {
    my $filename = shift;
    my $mode = shift;
    my ($name, $path, $suffix) = fileparse($filename);

    system("mkdir -p $path");

    open(FH, "$mode", "$filename")
	or die "cannot open $filename (mode=$mode): $!";

    print FH @_
	or die "write($filename) failed: $!";

    close FH
	or die "close($filename) failed: $!";
}

die "output-dir not specified"
    unless $output_dir;

die "expecting output_dir '$output_dir' to begin with '/tmp'"
    unless $output_dir =~ m!^/tmp!;

die "crt-file not specified"
    unless $crt_file;

# clean up successive runs for the same output directory.
system("rm -rf $output_dir");

write_to_file("$output_dir/conf/error-page-503.http", ">", <<EOF
HTTP/1.0 503 Service Unavailable
Pragma: no-cache
Cache-Control: private, max-age=0, no-cache, no-store
Connection: close
Content-Type: text/html

<html>
  <body>
  503
  </body>
</html>
EOF
    );

# Standard OpenShift haproxy preamble.
write_to_file("$output_dir/conf/haproxy.config", ">", <<EOF
global
  maxconn 20000
  daemon
  log stdout format raw local0
  nbthread 4
  tune.maxrewrite 8192
  tune.bufsize 32768
  stats socket $output_dir/haproxy.sock mode 600 level admin expose-fd listeners
  ssl-default-bind-options ssl-min-ver TLSv1.2

defaults
  log global
  option httplog
  option logasap
  errorfile 503 $output_dir/conf/error-page-503.http
  timeout connect 30s
  timeout client 30s
  timeout client-fin 1s
  timeout server 30s
  timeout server-fin 1s
  timeout http-request 10s
  timeout http-keep-alive 300s
  timeout tunnel 30s

frontend public
  bind :$public_http_port
  mode http
  tcp-request inspect-delay 5s
  tcp-request content accept if HTTP
  monitor-uri /_______internal_router_healthz

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]

  # check if we need to redirect/force using https.
  acl secure_redirect base,map_reg(os_route_http_redirect.map) -m found
  redirect scheme https if secure_redirect

  use_backend %[base,map_reg(os_http_be.map)]

  default_backend openshift_default

frontend public_ssl
  option tcplog
  bind :$public_https_port
  tcp-request inspect-delay 5s
  tcp-request content accept if { req_ssl_hello_type 1 }

  # if the connection is SNI and the route is a passthrough don't use the termination backend, just use the tcp backend
  # for the SNI case, we also need to compare it in case-insensitive mode (by converting it to lowercase) as RFC 4343 says
  acl sni req.ssl_sni -m found
  acl sni_passthrough req.ssl_sni,lower,map_reg($output_dir/conf/os_sni_passthrough.map) -m found
  use_backend %[req.ssl_sni,lower,map_reg($output_dir/conf/os_tcp_be.map)] if sni sni_passthrough

  # if the route is SNI and NOT passthrough enter the termination flow
  use_backend be_sni if sni

  # non SNI requests should enter a default termination backend rather than the custom cert SNI backend since it
  # will not be able to match a cert to an SNI host
  default_backend be_no_sni

##########################################################################
# TLS SNI
#
# When using SNI we can terminate encryption with custom certificates.
# Certs will be stored in a directory and will be matched with the SNI host header
# which must exist in the CN of the certificate.  Certificates must be concatenated
# as a single file (handled by the plugin writer) per the haproxy documentation.
#
# Finally, check re-encryption settings and re-encrypt or just pass along the unencrypted
# traffic
##########################################################################
backend be_sni
  server fe_sni :${fe_sni_port} weight 1

frontend fe_sni
  # terminate ssl on edge
  bind :${fe_sni_port} ssl crt $crt_file crt-list $output_dir/conf/cert_config.map
  mode http

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]

  # Search from most specific to general path (host case).
  # Note: If no match, haproxy uses the default_backend, no other
  #       use_backend directives below this will be processed.
  use_backend %[base,map_reg(os_edge_reencrypt_be.map)]

  default_backend openshift_default

# backend for when sni does not exist, or ssl term needs to happen on the edge
backend be_no_sni
  server fe_no_sni 127.0.0.1:${be_no_sni_port} weight 1

frontend fe_no_sni
  # terminate ssl on edge
  bind 127.0.0.1:${be_no_sni_port} ssl crt $crt_file
  mode http

  # Strip off Proxy headers to prevent HTTpoxy (https://httpoxy.org/)
  http-request del-header Proxy

  # DNS labels are case insensitive (RFC 4343), we need to convert the hostname into lowercase
  # before matching, or any requests containing uppercase characters will never match.
  http-request set-header Host %[req.hdr(Host),lower]

  # map to backend
  # Search from most specific to general path (host case).
  # Note: If no match, haproxy uses the default_backend, no other
  #       use_backend directives below this will be processed.
  use_backend %[base,map_reg($output_dir/conf/os_edge_reencrypt_be.map)]

  default_backend openshift_default

backend openshift_default
  mode http
  option forwardfor
  #option http-keep-alive
  option http-pretend-keepalive
EOF
    );

write_to_file("$output_dir/conf/cert_config.map", ">", "");
write_to_file("$output_dir/conf/os_http_be.map", ">", "");
write_to_file("$output_dir/conf/os_route_http_redirect.map", ">", "");
write_to_file("$output_dir/conf/os_edge_reencrypt_be.map", ">", "");
write_to_file("$output_dir/conf/os_sni_passthrough.map", ">", "");
write_to_file("$output_dir/conf/os_tcp_be.map", ">", "");

my @proxy;
my @os_edge_reencrypt_be;

for (my $i = 0; $i < $proxies; $i++) {
    my $proxy_name = "be_${route_type}:${namespace}:app-${i}-${route_type}";
    my $proxy_name_hash = md5_hex("$proxy_name");

    push(@proxy, <<EOF

backend ${proxy_name}
  mode http
  option redispatch
  option forwardfor
  balance ${balance_algorithm}

  timeout check 5000ms
  http-request set-header X-Forwarded-Host %[req.hdr(host)]
  http-request set-header X-Forwarded-Port %[dst_port]
  http-request set-header X-Forwarded-Proto http if !{ ssl_fc }
  http-request set-header X-Forwarded-Proto https if { ssl_fc }
  http-request set-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }
  http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)];proto-version=%[req.hdr(X-Forwarded-Proto-Version)]
  cookie ${proxy_name_hash} insert indirect nocache httponly
EOF
	);

    for (my $j = 0; $j < $servers; $j++) {
	my $endpoint = "pod:app-$i:replica-$j:10.128.0.$j:100${j}";
	my $endpoint_hash = md5_hex("$endpoint");
	push(@proxy, "server $endpoint ${target_host}:${target_port} cookie $endpoint_hash weight $weight\n");
    }

    push(@os_edge_reencrypt_be, "^app-${i}-${route_type}\$ $proxy_name\n");
}

write_to_file("$output_dir/conf/haproxy.config", ">>", "@proxy");
write_to_file("$output_dir/conf/os_edge_reencrypt_be.map", ">>", "@os_edge_reencrypt_be");
