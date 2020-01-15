Understand how reload works with dataplaneapi

# Build and run

    $ make
```console
[NOTICE] 014/181232 (1) : New program 'api' (8) forked
[NOTICE] 014/181232 (1) : New worker #1 (9) forked
time="2020-01-15T18:12:32Z" level=info msg="HAProxy Data Plane API v1.2.4 2622bb6.dev"
time="2020-01-15T18:12:32Z" level=info msg="Build from: https://github.com/haproxytech/dataplaneapi.git"
time="2020-01-15T18:12:32Z" level=info msg="Build date: 2020-01-15T12:43:21"
time="2020-01-15T18:12:32Z" level=info msg="Serving data plane at http://[::]:5555"
```
	
# Add some new frontends

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

# Look at the debug and process table
```console
[NOTICE] 014/181232 (1) : New worker #1 (9) forked
time="2020-01-15T18:12:32Z" level=info msg="HAProxy Data Plane API v1.2.4 2622bb6.dev"
time="2020-01-15T18:12:32Z" level=info msg="Build from: https://github.com/haproxytech/dataplaneapi.git"
time="2020-01-15T18:12:32Z" level=info msg="Build date: 2020-01-15T12:43:21"
time="2020-01-15T18:12:32Z" level=info msg="Serving data plane at http://[::]:5555"
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55466" request="/v2/services/haproxy/configuration/frontends?version=1"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=31.260405ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55470" request="/v2/services/haproxy/configuration/frontends?version=2"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=33.505257ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55474" request="/v2/services/haproxy/configuration/frontends?version=3"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=27.097778ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55478" request="/v2/services/haproxy/configuration/frontends?version=4"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=29.386069ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55484" request="/v2/services/haproxy/configuration/frontends?version=5"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=29.226821ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55488" request="/v2/services/haproxy/configuration/frontends?version=6"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=31.109746ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55492" request="/v2/services/haproxy/configuration/frontends?version=7"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=32.53628ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55496" request="/v2/services/haproxy/configuration/frontends?version=8"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=31.882176ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55500" request="/v2/services/haproxy/configuration/frontends?version=9"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=80B status=202 took=30.104003ms
time="2020-01-15T18:12:55Z" level=info msg="started handling request" method=POST remote="172.17.0.1:55504" request="/v2/services/haproxy/configuration/frontends?version=10"
time="2020-01-15T18:12:55Z" level=info msg="completed handling request" length=81B status=202 took=30.343433ms
time="2020-01-15T18:12:57Z" level=debug msg="Reload started..."
time="2020-01-15T18:12:57Z" level=debug msg="Reload finished."
time="2020-01-15T18:12:57Z" level=debug msg="Time elapsed: 1.34355ms"
time="2020-01-15T18:12:57Z" level=debug msg="Reload successful"
[WARNING] 014/181257 (1) : Reexecuting Master process
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_1' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_10' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_2' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_3' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_4' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_5' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_6' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_7' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_8' has no 'bind' directive. Please declare it as a backend if this was intended.
[WARNING] 014/181257 (1) : config : frontend 'test_frontend_9' has no 'bind' directive. Please declare it as a backend if this was intended.
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
00000000:GLOBAL.accept(0006)=0021 from [unix:1] ALPN=<none>
00000000:GLOBAL.srvcls[adfd:ffffffff]
00000000:GLOBAL.clicls[adfd:ffffffff]
00000000:GLOBAL.closed[adfd:ffffffff]
[WARNING] 014/181257 (9) : Stopping frontend GLOBAL in 0 ms.
[WARNING] 014/181257 (9) : Stopping frontend FE in 0 ms.
[WARNING] 014/181257 (9) : Stopping backend static in 0 ms.
time="2020-01-15T18:12:57Z" level=info msg="Reloaded Data Plane API"
[WARNING] 014/181257 (9) : Stopping backend app in 0 ms.
[WARNING] 014/181257 (9) : Proxy GLOBAL stopped (FE: 1 conns, BE: 1 conns).
[WARNING] 014/181257 (9) : Proxy FE stopped (FE: 0 conns, BE: 0 conns).
[WARNING] 014/181257 (9) : Proxy static stopped (FE: 0 conns, BE: 0 conns).
[WARNING] 014/181257 (9) : Proxy app stopped (FE: 0 conns, BE: 0 conns).
[NOTICE] 014/181257 (1) : New worker #1 (42) forked
[WARNING] 014/181258 (1) : Former worker #1 (9) exited with code 0 (Exit)
```
