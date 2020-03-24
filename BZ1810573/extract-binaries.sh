#!/bin/bash

set -x

topdir="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd -P)"

grab_haproxy_bin() {
    local rpmfile=$(basename "$1")
    [ -f $rpmfile ] || wget $1
    local rpmdir=$(basename "$rpmfile" .rpm)
    mkdir -p $rpmdir
    rpmpeek $rpmfile cp -rv ./usr/sbin/haproxy $topdir/$rpmdir
    rm -f $rpmfile
}

grab_haproxy_bin http://download.eng.bos.redhat.com/brewroot/vol/rhel-7/packages/haproxy/1.8.17/3.el7/x86_64/haproxy18-1.8.17-3.el7.x86_64.rpm
grab_haproxy_bin https://people.redhat.com/kwalker/repos/bz1810573/x86_64/Packages/haproxy18-1.8.17-4.el7.x86_64.rpm
grab_haproxy_bin https://github.com/frobware/haproxy-hacks/raw/master/RPMs/haproxy18-1.8.23-2.el7.x86_64.rpm
