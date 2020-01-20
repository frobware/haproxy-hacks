# Understand how reload works with dataplaneapi

## Build and run

    $ make
```console
[NOTICE] 014/181232 (1) : New program 'api' (8) forked
[NOTICE] 014/181232 (1) : New worker #1 (9) forked
time="2020-01-15T18:12:32Z" level=info msg="HAProxy Data Plane API v1.2.4 2622bb6.dev"
time="2020-01-15T18:12:32Z" level=info msg="Build from: https://github.com/haproxytech/dataplaneapi.git"
time="2020-01-15T18:12:32Z" level=info msg="Build date: 2020-01-15T12:43:21"
time="2020-01-15T18:12:32Z" level=info msg="Serving data plane at http://[::]:5555"
```
	
## Add some new frontends

    $ ./add-front-end
```console	
{
  "default_backend": "app",
  "maxconn": 2000,
  "mode": "http",
  "name": "test_frontend_1"
}
{
  "default_backend": "app",
  "maxconn": 2000,
  "mode": "http",
  "name": "test_frontend_2"
}
...
```

## Look at the debug output

```console
[NOTICE] 014/182229 (1) : New program 'api' (8) forked
[NOTICE] 014/182229 (1) : New worker #1 (9) forked
time="2020-01-15T18:22:29Z" level=info msg="HAProxy Data Plane API v1.2.4 2622bb6.dev"
time="2020-01-15T18:22:29Z" level=info msg="Build from: https://github.com/haproxytech/dataplaneapi.git"
time="2020-01-15T18:22:29Z" level=info msg="Build date: 2020-01-15T12:43:21"
time="2020-01-15T18:22:29Z" level=info msg="Serving data plane at http://[::]:5555"
^C[WARNING] 014/182757 (1) : Exiting Master process...
time="2020-01-15T18:27:57Z" level=info msg="Shutting down... "
time="2020-01-15T18:27:57Z" level=info msg="Stopped serving data plane at http://[::]:5555"
[ALERT] 014/182757 (1) : Current worker #1 (9) exited with code 130 (Interrupt)
[ALERT] 014/182757 (1) : Current program 'api' (8) exited with code 0 (Exit)
[WARNING] 014/182757 (1) : All workers exited. Exiting... (130)
make: *** [Makefile:4: run] Error 130

aim@spicy:~/frobware/haproxy-hacks/dataplane-api
$ make
docker build -t dataplane-api-test . && \
docker run -p 5555:5555 dataplane-api-test
Sending build context to Docker daemon 22.18 MB
Step 1/9 : FROM centos:7
 ---> 5e35e350aded
Step 2/9 : RUN yum install -y psmisc procps-ng rsyslog sysvinit-tools socat
 ---> Using cache
 ---> 1150af12fa6f
Step 3/9 : RUN rpm -ivh http://spicy.frobware.com/~aim/x86_64/haproxy20-2.0.12-1.el7.x86_64.rpm
 ---> Using cache
 ---> 7705b7ce4ee5
Step 4/9 : RUN mkdir -p /var/lib/haproxy &&     mkdir -p /var/lib/haproxy/run &&     mkdir -p /var/lib/haproxy/router/{certs,cacerts,whitelists} &&     mkdir -p /var/lib/haproxy/{conf/.tmp,run,bin,log} &&     touch /var/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}
 ---> Using cache
 ---> 4840893732bd
Step 5/9 : COPY dataplaneapi /usr/local/bin
 ---> Using cache
 ---> de00026c5b7e
Step 6/9 : COPY reload-haproxy /usr/local/bin
 ---> Using cache
 ---> 7db5cae4bd3e
Step 7/9 : COPY haproxy.cfg /etc/haproxy/haproxy.cfg
 ---> Using cache
 ---> 3d29a1421729
Step 8/9 : EXPOSE 5555
 ---> Using cache
 ---> 72a824507b64
Step 9/9 : ENTRYPOINT /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d
 ---> Using cache
 ---> 38c7915f0edc
Successfully built 38c7915f0edc
Note: setting global.maxconn to 524264.
Available polling systems :
      epoll : pref=300,  test result OK
       poll : pref=200,  test result OK
     select : pref=150,  test result FAILED
Total: 3 (2 usable), will use epoll.

Available filters :
	[SPOE] spoe
	[COMP] compression
	[CACHE] cache
	[TRACE] trace
Using epoll() as the polling mechanism.
[NOTICE] 014/182758 (1) : New program 'api' (8) forked
[NOTICE] 014/182758 (1) : New worker #1 (9) forked
time="2020-01-15T18:27:59Z" level=info msg="HAProxy Data Plane API v1.2.4 2622bb6.dev"
time="2020-01-15T18:27:59Z" level=info msg="Build from: https://github.com/haproxytech/dataplaneapi.git"
time="2020-01-15T18:27:59Z" level=info msg="Build date: 2020-01-15T12:43:21"
time="2020-01-15T18:27:59Z" level=info msg="Serving data plane at http://[::]:5555"
time="2020-01-15T18:28:34Z" level=info msg="started handling request" method=GET remote="172.17.0.1:33308" request=/v2/services/haproxy/configuration/frontends
time="2020-01-15T18:28:34Z" level=info msg="completed handling request" length=62B status=200 took="430.274µs"
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33348" request="/v2/services/haproxy/configuration/frontends?version=1"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=40.533917ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33352" request="/v2/services/haproxy/configuration/frontends?version=2"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=26.775422ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33356" request="/v2/services/haproxy/configuration/frontends?version=3"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=33.194243ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33360" request="/v2/services/haproxy/configuration/frontends?version=4"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=27.648286ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33364" request="/v2/services/haproxy/configuration/frontends?version=5"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=28.259927ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33368" request="/v2/services/haproxy/configuration/frontends?version=6"
time="2020-01-15T18:28:39Z" level=info msg="completed handling request" length=80B status=202 took=34.09415ms
time="2020-01-15T18:28:39Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33374" request="/v2/services/haproxy/configuration/frontends?version=7"
time="2020-01-15T18:28:40Z" level=info msg="completed handling request" length=80B status=202 took=29.697776ms
time="2020-01-15T18:28:40Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33378" request="/v2/services/haproxy/configuration/frontends?version=8"
time="2020-01-15T18:28:40Z" level=info msg="completed handling request" length=80B status=202 took=34.09474ms
time="2020-01-15T18:28:40Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33382" request="/v2/services/haproxy/configuration/frontends?version=9"
time="2020-01-15T18:28:40Z" level=info msg="completed handling request" length=80B status=202 took=30.777635ms
time="2020-01-15T18:28:40Z" level=info msg="started handling request" method=POST remote="172.17.0.1:33386" request="/v2/services/haproxy/configuration/frontends?version=10"
time="2020-01-15T18:28:40Z" level=info msg="completed handling request" length=81B status=202 took=30.625068ms
time="2020-01-15T18:28:44Z" level=debug msg="Reload started..."
time="2020-01-15T18:28:44Z" level=debug msg="Reload finished."
time="2020-01-15T18:28:44Z" level=debug msg="Time elapsed: 1.380798ms"
time="2020-01-15T18:28:44Z" level=debug msg="Reload successful"
[WARNING] 014/182844 (1) : Reexecuting Master process
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_1' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_10' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_2' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_3' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_4' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_5' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_6' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_7' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_8' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/182844 (1) : config : frontend 'test_frontend_9' has no 'bind' directive. Please declare it as a backend if this was intended.
Note: setting global.maxconn to 524264.
Available polling systems :
      epoll : pref=300,  test result OK
       poll : pref=200,  test result OK
     select : pref=150,  test result FAILED
Total: 3 (2 usable), will use epoll.

Available filters :
	[SPOE] spoe
	[COMP] compression
	[CACHE] cache
	[TRACE] trace
Using epoll() as the polling mechanism.
00000000:GLOBAL.accept(0006)=0020 from [unix:1] ALPN=<none>
00000000:GLOBAL.srvcls[adfd:ffffffff]
00000000:GLOBAL.clicls[adfd:ffffffff]
00000000:GLOBAL.closed[adfd:ffffffff]
[WARNING] 014/182844 (9) : Stopping frontend GLOBAL in 0 ms.
[WARNING] 014/182844 (9) : Stopping frontend FE in 0 ms.
[WARNING] 014/182844 (9) : Stopping backend static in 0 ms.
[WARNING] 014/182844 (9) : Stopping backend app in 0 ms.
time="2020-01-15T18:28:44Z" level=info msg="Reloaded Data Plane API"
[WARNING] 014/182844 (9) : Proxy GLOBAL stopped (FE: 1 conns, BE: 1 conns).
[WARNING] 014/182844 (9) : Proxy FE stopped (FE: 0 conns, BE: 0 conns).
[WARNING] 014/182844 (9) : Proxy static stopped (FE: 0 conns, BE: 0 conns).
[WARNING] 014/182844 (9) : Proxy app stopped (FE: 0 conns, BE: 0 conns).
[NOTICE] 014/182844 (1) : New worker #1 (57) forked
[WARNING] 014/182845 (1) : Former worker #1 (9) exited with code 0 (Exit)
```

## Process table before adding new frontends

```console
$ docker exec -it sharp_curie bash
[root@d598af39aca5 /]# ps -ax
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d
    8 ?        Sl     0:00 /usr/local/bin/dataplaneapi --host 0.0.0.0 --port 5555 --haproxy-bin /usr/sbin/haproxy --config-file /etc/haproxy/haproxy.cfg --reload-cmd kill -SI
    9 ?        Sl     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d
   29 ?        Ss     0:00 bash
   43 ?        R+     0:00 ps -ax
```

## Process table after adding new frontends

```console
[root@d598af39aca5 /]# ps -ax
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d -sf 8 9 -x /var/lib/haproxy/run/haproxy.sock
    8 ?        Sl     0:00 /usr/local/bin/dataplaneapi --host 0.0.0.0 --port 5555 --haproxy-bin /usr/sbin/haproxy --config-file /etc/haproxy/haproxy.cfg --reload-cmd kill -SI
   29 ?        Ss     0:00 bash
   57 ?        Sl     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d -sf 8 9 -x /var/lib/haproxy/run/haproxy.sock
   65 ?        R+     0:00 ps -ax
```

So, new pids.
