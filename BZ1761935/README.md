Notes to reproduce https://bugzilla.redhat.com/show_bug.cgi?id=1761935

## Build HAProxy locally; see https://github.com/frobware/haproxy-hacks/blob/master/BZ1690146/README.md

## Start up haproxy, reloading every second

    while :; do ~/haproxy-hacks/BZ1761935/reload-proxy ; sleep 1; done

### Start up our backend server

    node ~/haproxy-hacks/BZ1761935/server.js

## Apply some load

    hey -c 2 -m GET -q 1 -z 50s http://localhost:4242/ba

hey(1) comes from https://github.com/rakyll/hey

From the haproxy reload you should see:

```console
$ while :
> do
> ~/haproxy-hacks/BZ1761935/reload-proxy 
> sleep 1
> done
10103
10143
10158
10173
10188
10203
10229
10244
10259
```

The numbers are the old pids as the proxy is restarted.

From the node process you should see:

```console
$ node ~/haproxy-hacks/BZ1761935/server.js 
Server running at http://0.0.0.0:9000/
Server running at http://0.0.0.0:9001/
Server running at http://0.0.0.0:9002/
{ host: 'localhost' }
{ host: 'localhost' }
{ host: 'localhost' }
{ host: 'localhost' }
{ host: 'localhost' }
```

When the hey(1) load is applied you will occasionally see the
"Connection: close" header injected into the backend request:


```console

{ 'user-agent': 'hey/0.0.1',
  'content-type': 'text/html',
  'accept-encoding': 'gzip',
  host: 'localhost:4242',
  'x-forwarded-port': '4242',
  'x-forwarded-for': '127.0.0.1' }
{ 'user-agent': 'hey/0.0.1',
  'content-type': 'text/html',
  'accept-encoding': 'gzip',
  host: 'localhost:4242',
  'x-forwarded-port': '4242',
  'x-forwarded-for': '127.0.0.1' }
{ host: 'localhost' }
{ 'user-agent': 'hey/0.0.1',
  'content-type': 'text/html',
  'accept-encoding': 'gzip',
  host: 'localhost:4242',
  'x-forwarded-port': '4242',
  'x-forwarded-for': '127.0.0.1',
  connection: 'close' }
```

If your remove the continuous reload of the proxy then I never see
the "connection: close" header injected.

## Status values *without* reloading

```console
$ hey -c 2 -m GET -q 1 -z 30s http://localhost:4242/bar
Summary:
  Total:	30.0064 secs
  Slowest:	0.0053 secs
  Fastest:	0.0012 secs
  Average:	0.0028 secs
  Requests/sec:	1.9996
  

Response time histogram:
  0.001 [1]	|■■■
  0.002 [5]	|■■■■■■■■■■■■■
  0.002 [4]	|■■■■■■■■■■
  0.002 [7]	|■■■■■■■■■■■■■■■■■■
  0.003 [16]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.003 [8]	|■■■■■■■■■■■■■■■■■■■■
  0.004 [13]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.004 [4]	|■■■■■■■■■■
  0.004 [1]	|■■■
  0.005 [0]	|
  0.005 [1]	|■■■


Latency distribution:
  10% in 0.0016 secs
  25% in 0.0023 secs
  50% in 0.0027 secs
  75% in 0.0033 secs
  90% in 0.0038 secs
  95% in 0.0040 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0012 secs, 0.0053 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0008 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0001 secs
  resp wait:	0.0025 secs, 0.0010 secs, 0.0037 secs
  resp read:	0.0001 secs, 0.0000 secs, 0.0002 secs

Status code distribution:
  [200]	60 responses
```

## Status responses *with* reload taking place

```console
$ hey -c 2 -m GET -q 1 -z 30s http://localhost:4242/bar

Summary:
  Total:	30.0081 secs
  Slowest:	0.0055 secs
  Fastest:	0.0015 secs
  Average:	0.0034 secs
  Requests/sec:	1.9995
  

Response time histogram:
  0.001 [1]	|■■■■
  0.002 [4]	|■■■■■■■■■■■■■■■
  0.002 [3]	|■■■■■■■■■■■
  0.003 [5]	|■■■■■■■■■■■■■■■■■■
  0.003 [9]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.003 [8]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.004 [11]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.004 [10]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.005 [5]	|■■■■■■■■■■■■■■■■■■
  0.005 [2]	|■■■■■■■
  0.005 [2]	|■■■■■■■


Latency distribution:
  10% in 0.0022 secs
  25% in 0.0028 secs
  50% in 0.0035 secs
  75% in 0.0041 secs
  90% in 0.0044 secs
  95% in 0.0049 secs
  0% in 0.0000 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0005 secs, 0.0015 secs, 0.0055 secs
  DNS-lookup:	0.0002 secs, 0.0000 secs, 0.0010 secs
  req write:	0.0001 secs, 0.0000 secs, 0.0002 secs
  resp wait:	0.0026 secs, 0.0013 secs, 0.0051 secs
  resp read:	0.0001 secs, 0.0001 secs, 0.0004 secs

Status code distribution:
  [200]	60 responses
```
