#!/bin/bash

set -eu

o="/tmp/haproxy-info-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$o"

pushd "$o"

: ${N:=15}
: ${INTERVAL:=5}

# sh-4.4$ echo "help" | socat /var/lib/haproxy/run/haproxy.sock stdio 2>&1 |grep show
#   show tls-keys [id|*]: show tls keys references or dump tls ticket keys when id specified
#   show sess [id] : report the list of current sessions or dump this session
#   show info      : report information about the running process [json|typed]
#   show stat      : report counters for each proxy and server [json|typed]
#   show schema json : report schema used for stats
#   show resolvers [id]: dumps counters from all resolvers section and
#   show table [id]: report table usage stats or dump this table's contents
#   show peers [peers section]: dump some information about all the peers or this peers section
#   show servers state [id]: dump volatile server information (for backend <id>)
#   show backend   : list backends in the current running config
#   show errors    : report last request and response errors for each proxy
#   show env [var] : dump environment variables known to the process
#   show cli sockets : dump list of cli sockets
#   show cli level   : display the level of the current CLI session
#   show fd [num] : dump list of file descriptors in use
#   show activity : show per-thread activity stats (for support/developers)
#   show startup-logs : report logs emitted during HAProxy startup
#   show cache     : show cache status
#   show acl [id]  : report available acls or dump an acl's contents
#   show map [id]  : report available maps or dump a map's contents
#   show pools     : report information about the memory pools usage
#   show profiling : show CPU profiling options
#   show threads   : show some threads debugging information

for i in $(seq 1 $N)
do
    d="$(date +%H%M%S)"
    mkdir -p "$d/$i"
    ps -aux > "$d/$i/ps"
    ss -tun -a sport = :80 > "$d/$i/ss-80"
    ss -tun -a sport = :443 > "$d/$i/ss-443"
    ss -aton > "$d/$i/ss"
    for j in info stat sess errors peers activity cache backend "servers state"
    do
	echo "show $j" | socat /var/lib/haproxy/run/haproxy.sock stdio > "$d/$i/${j}.log"
    done
    echo "$i/$N: sleeping for ${INTERVAL}s"
    sleep "$INTERVAL"
done
