## Build haproxy without optimisation (aka a debug build)

    rm -rf ~/rpmbuild/RPMS

    rpmbuild --define '_build_type debug' -ba --noclean haproxy.spec

    sudo rpm -Uvh --force $(find $HOME/rpmbuild/RPMS/x86_64 -name 'haproxy*.rpm')
    sudo rpm -ev haproxy28 haproxy28-debuginfo haproxy-debugsource

Optimised builds:

    rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/rhaos-4.17-rhel-9/haproxy-debugsource-2.8.10-1.rhaos4.17.el9.x86_64.rpm
    rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/rhaos-4.17-rhel-9/haproxy28-2.8.10-1.rhaos4.17.el9.x86_64.rpm
    rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/rhaos-4.17-rhel-9/haproxy28-debuginfo-2.8.10-1.rhaos4.17.el9.x86_64.rpm

Debug builds:

     rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/debug/rhaos-4.17-rhel-9/haproxy-debugsource-2.8.10-1.rhaos4.17.el9.x86_64.rpm
     rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/debug/rhaos-4.17-rhel-9/haproxy28-2.8.10-1.rhaos4.17.el9.x86_64.rpm
     rpm -Uvh https://github.com/frobware/haproxy-builds/raw/master/debug/rhaos-4.17-rhel-9/haproxy28-debuginfo-2.8.10-1.rhaos4.17.el9.x86_64.rpm

https://brewweb.engineering.redhat.com/brew/taskinfo?taskID=62236176
