# Notes from a number of shells...

This captures some debug runes when trying to reproduce:

   https://bugzilla.redhat.com/show_bug.cgi?id=1690146


This bug shows that we're using haproxy-1.8.17, let's get that:

```sh
git clone http://git.haproxy.org/git/haproxy-1.8.git/
cd haproxy-1.18
git checkout v1.18.17
```

To find out how this binary is being built:

```sh
mkdir /tmp/haproxy
cd /tmp/haxproy
yumdownloader --source haproxy
haproxy-1.8.21-1.fc30.src.rpm
rpm2cpio haproxy-1.8.21-1.fc30.src.rpm | cpio -idmv
```

Looking in the `haproxy.spec` file I see:

```
%{__make} %{?_smp_mflags} CPU="generic" TARGET="linux2628"
USE_OPENSSL=1 USE_PCRE=1 USE_ZLIB=1 USE_LUA=1 USE_CRYPT_H=1
USE_SYSTEMD=1 USE_LINUX_TPROXY=1 USE_GETADDRINFO=1 ${regparm_opts}
ADDINC="%{optflags}" ADDLIB="%{__global_ldflags}"
```

Build our binary with [debug](0001-Build-with-debug.patch)

```
cd ~/haproxy-1.8
git apply ~/haproxy-hacks/BZ1690146/0001-Build-with-debug.patch
make CPU="generic" TARGET="linux2628" USE_OPENSSL=1 USE_PCRE=1 \
	USE_ZLIB=1 USE_LUA=1 USE_CRYPT_H=1 USE_SYSTEMD=1 USE_LINUX_TPROXY=1 USE_GETADDRINFO=1
./haxproy -V
HA-Proxy version 1.8.17 2019/01/08
Copyright 2000-2019 Willy Tarreau <willy@haproxy.org>
```

# Test strategy

I wanted to stress test reloading the proxy while applying some load.

## Run the [backend](server.js):

```sh
$ node ~/haproxy-hacks/BZ1690146/server.js &
```

## Run [haproxy](./reload-proxy)

```sh
$ ~/haproxy-hacks/BZ1690146/reload-proxy
```

## Run apache benchmark and verify everything is working:

```console
$ ab -v 2 -c 100 -n 10000000 -k http://localhost:4242/

LOG: header received:
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Fri, 27 Sep 2019 16:32:17 GMT
Connection: close

{"connection":"Keep-Alive","user-agent":"ApacheBench/2.3","accept":"*/*","host":"localhost:4242","x-forwarded-port":"4242","x-forwarded-for":"127.0.0.1"}
There's no place like 0.0.0.0:9001

LOG: header received:
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Fri, 27 Sep 2019 16:32:17 GMT
Connection: close

{"connection":"Keep-Alive","user-agent":"ApacheBench/2.3","accept":"*/*","host":"localhost:4242","x-forwarded-port":"4242","x-forwarded-for":"127.0.0.1"}
There's no place like 0.0.0.0:9000
```

Now that this is working I tried sitting in a loop reloading ad nausem
to see if I could reproduce the behaviour observed in the bug where
there are lingering processes that are still listening:

```console
$ while :; do ./reload-haproxy; sleep 0.1; done | ts
Sep 27 17:39:01 14281 14236 14216 14174 12763 11185 10884 10538                                                                           │haproxy 9807  aim   72u  IPv4 12464860      0t0  TCP localhost:4242->localhost:45212 (ESTABLISHED)
Sep 27 17:39:02 14362 14281 14236 14216 14174 12763 11185 10884 10538                                                                     │haproxy 9807  aim   86u  IPv4 12464862      0t0  TCP localhost:4242->localhost:45216 (ESTABLISHED)
Sep 27 17:39:02 14401 14362 12763 11185 10884 10538                                                                                       │haproxy 9807  aim  130u  IPv4 12464843      0t0  TCP localhost:4242->localhost:45028 (ESTABLISHED)
Sep 27 17:39:02 14449 14401 14362 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  143u  IPv4 12471716      0t0  TCP localhost:4242->localhost:44806 (ESTABLISHED)
Sep 27 17:39:02 14491 14449 12763 11185 10884 10538                                                                                       │haproxy 9807  aim  144u  IPv4 12471717      0t0  TCP localhost:4242->localhost:44808 (ESTABLISHED)
Sep 27 17:39:02 14531 14491 14449 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  145u  IPv4 12471718      0t0  TCP localhost:4242->localhost:44810 (ESTABLISHED)
Sep 27 17:39:02 14575 14531 14491 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  146u  IPv4 12471719      0t0  TCP localhost:4242->localhost:44812 (ESTABLISHED)
Sep 27 17:39:02 14617 14575 14531 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  147u  IPv4 12471720      0t0  TCP localhost:4242->localhost:44814 (ESTABLISHED)
Sep 27 17:39:03 14660 14617 14575 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  148u  IPv4 12471721      0t0  TCP localhost:4242->localhost:44816 (ESTABLISHED)
Sep 27 17:39:03 14703 14660 14617 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  149u  IPv4 12471722      0t0  TCP localhost:4242->localhost:EtherNet/IP-2 (ESTABLISHED)
Sep 27 17:39:03 14745 14703 12763 11185 10884 10538                                                                                       │haproxy 9807  aim  153u  IPv4 12471726      0t0  TCP localhost:4242->localhost:44826 (ESTABLISHED)
[WARNING] 269/173903 (14762) : Failed to get the number of sockets to be transferred !                                                    │haproxy 9807  aim  154u  IPv4 12471727      0t0  TCP localhost:4242->localhost:44828 (ESTABLISHED)
[ALERT] 269/173903 (14762) : Failed to get the sockets from the old process!                                                              │haproxy 9807  aim  155u  IPv4 12471728      0t0  TCP localhost:4242->localhost:44830 (ESTABLISHED)
Sep 27 17:39:03 14745 14703 12763 11185 10884 10538                                                                                       │haproxy 9807  aim  159u  IPv4 12471732      0t0  TCP localhost:4242->localhost:44838 (ESTABLISHED)
Sep 27 17:39:03 14799 14745 14703 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  169u  IPv4 12471742      0t0  TCP localhost:4242->localhost:44858 (ESTABLISHED)
Sep 27 17:39:03 14841 14799 14745 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  173u  IPv4 12471746      0t0  TCP localhost:4242->localhost:44866 (ESTABLISHED)
Sep 27 17:39:03 14881 14841 14799 14745 12763 11185 10884 10538                                                                           │haproxy 9807  aim  187u  IPv4 12471749      0t0  TCP localhost:4242->localhost:44872 (ESTABLISHED)
Sep 27 17:39:03 14925 14881 14841 14799 12763 11185 10884 10538                                                                           │haproxy 9807  aim  188u  IPv4 12471750      0t0  TCP localhost:4242->localhost:44874 (ESTABLISHED)
Sep 27 17:39:04 14952 14925 14881 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  190u  IPv4 12471752      0t0  TCP localhost:4242->localhost:44878 (ESTABLISHED)
Sep 27 17:39:04 14997 14952 14925 14881 12763 11185 10884 10538                                                                           │haproxy 9807  aim  191u  IPv4 12471753      0t0  TCP localhost:4242->localhost:44880 (ESTABLISHED)
Sep 27 17:39:04 15053 14997 14952 14925 12763 11185 10884 10538                                                                           │haproxy 9807  aim  193u  IPv4 12471755      0t0  TCP localhost:4242->localhost:44884 (ESTABLISHED)
Sep 27 17:39:04 15085 15053 14997 14925 12763 11185 10884 10538                                                                           │haproxy 9807  aim  194u  IPv4 12471756      0t0  TCP localhost:4242->localhost:44886 (ESTABLISHED)
[WARNING] 269/173904 (15134) : We didn't get the expected number of sockets (expecting 2 got 1)                                           │haproxy 9807  aim  197u  IPv4 12471759      0t0  TCP localhost:4242->localhost:44892 (ESTABLISHED)
[ALERT] 269/173904 (15134) : Failed to get the sockets from the old process!                                                              │haproxy 9807  aim  198u  IPv4 12471760      0t0  TCP localhost:4242->localhost:44894 (ESTABLISHED)
Sep 27 17:39:04 15085 15053 14925 12763 11185 10884 10538                                                                                 │haproxy 9807  aim  233u  IPv4 12466072      0t0  TCP localhost:4242->localhost:45176 (ESTABLISHED)
Sep 27 17:39:04 15147 15085 15053 14925 12763 11185 10884 10538                                                                           │aim@spicy:~/haproxy-1.8
Sep 27 17:39:04 15188 15147 15085 14925 12763 11185 10884 10538                                                                           │$
Sep 27 17:39:04 15222 15188 15147 14925 12763 11185 10884 10538                                                                           │aim@spicy:~/haproxy-1.8
Sep 27 17:39:04 15260 15222 15188 14925 12763 11185 10884 10538
```

And all of the time I have this running in a different shell:

```console
$ ab -v 2 -c 100 -n 10000000 -k http://localhost:4242/ | grep '^HTTP' |grep -v 200
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
HTTP/1.0 504 Gateway Time-out
```

And looking at the number of processes listening on port 4242 I see:

```console
$ lsof -i:4242 | wc -l
69
```

This seems to go up/down:

```console
$ lsof -i:4242 | wc -l
20
```

At some point we see:

```console
$ ps -ef |grep haproxy
aim      15486  1350 99 17:39 ?        00:50:08 /home/aim/haproxy-1.8/haproxy -f /home/aim/haproxy-hacks/BZ1690146/haproxy.cfg -p /var/tmp/haproxy/run/haproxy.pid -x /var/tmp/haproxy/run/haproxy.sock -sf 15449 15407 15366 15284 15260 14925 12763 11185 10884 10538
```

And looking at the pid tree for `15486` I see:

```console
$ pstree  -alp -A 15486
haproxy,15486 -f /home/aim/haproxy-hacks/BZ1690146/haproxy.cfg -p /var/tmp/haproxy/run/haproxy.pid -x /var/tmp/haproxy/run/haproxy.sock -sf 15449 15407 15366 15284 15260 14925 12763 11185 10884 10538
  |-{haproxy},15487
  |-{haproxy},15488
  |-{haproxy},15489
  |-{haproxy},15491
  |-{haproxy},15493
  |-{haproxy},15503
  |-{haproxy},15505
  |-{haproxy},15506
  |-{haproxy},15507
  |-{haproxy},15508
  |-{haproxy},15509
  |-{haproxy},15510
  |-{haproxy},15511
  |-{haproxy},15512
  |-{haproxy},15513
  |-{haproxy},15514
  |-{haproxy},15515
  |-{haproxy},15516
  |-{haproxy},15517
  |-{haproxy},15518
  |-{haproxy},15519
  |-{haproxy},15520
  |-{haproxy},15521
  |-{haproxy},15522
  |-{haproxy},15523
  |-{haproxy},15526
  |-{haproxy},15529
  |-{haproxy},15531
  `-{haproxy},15532
```

The soft-reset identifies the current pids:

```console
$ for i in 15449 15407 15366 15284 15260 14925 12763 11185 10884 10538
> do
> ps -fp $i
> done
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
UID        PID  PPID  C STIME TTY          TIME CMD
```

and all of these have disappeared - ergo, no lingering processes. ¯\_(ツ)_/¯
