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
