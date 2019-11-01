# timeout tunnel configuration in haproxy

I misunderstood what `timeout tunnel 1h` actually meant. I thought,
irrespective of active traffic, any old (post reloading) haproxy
processes would be terminated after 1h. Whereas it's the opposite. If
after $DURATION there has been _zero_ activity then, and only then,
can the process be terminated.

## Start up haproxy, reloading every second

```console
$ while :; do ~/haproxy-hacks/BZ1743291/reload-proxy ; sleep 5; done
Thu 31 Oct 2019 05:45:00 PM GMT -- oldpids: 
Thu 31 Oct 2019 05:45:05 PM GMT -- oldpids: 8555
Thu 31 Oct 2019 05:45:10 PM GMT -- oldpids: 8574
Thu 31 Oct 2019 05:45:15 PM GMT -- oldpids: 8592
Thu 31 Oct 2019 05:45:20 PM GMT -- oldpids: 8610
Thu 31 Oct 2019 05:45:25 PM GMT -- oldpids: 8632
Thu 31 Oct 2019 05:45:30 PM GMT -- oldpids: 8651
Thu 31 Oct 2019 05:45:35 PM GMT -- oldpids: 8675
Thu 31 Oct 2019 05:45:40 PM GMT -- oldpids: 8693
Thu 31 Oct 2019 05:45:45 PM GMT -- oldpids: 8716
Thu 31 Oct 2019 05:45:51 PM GMT -- oldpids: 8734
Thu 31 Oct 2019 05:45:56 PM GMT -- oldpids: 8755
Thu 31 Oct 2019 05:46:01 PM GMT -- oldpids: 8776
Thu 31 Oct 2019 05:46:06 PM GMT -- oldpids: 8794
Thu 31 Oct 2019 05:46:11 PM GMT -- oldpids: 8815
Thu 31 Oct 2019 05:46:16 PM GMT -- oldpids: 8834
Thu 31 Oct 2019 05:46:21 PM GMT -- oldpids: 8853
```

## Run the backend

	go run server.go

## Make some websocket connections

    go run client.go

Observe traffic hitting the server:

```console
2019/10/31 17:47:43.171279 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.272800 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.374369 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.475807 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.577198 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.678700 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.780846 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.882253 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:43.984263 server.go:42: recv message from "127.0.0.1:36728"
2019/10/31 17:47:44.085783 server.go:42: recv message from "127.0.0.1:36728"
```

And here pid `20369` is a constant as that has established connections
between our server and client.

```console
Fri 01 Nov 2019 08:29:55 AM GMT -- oldpids: 20369 2279
Fri 01 Nov 2019 08:30:00 AM GMT -- oldpids: 20369 2304
Fri 01 Nov 2019 08:30:05 AM GMT -- oldpids: 20369 2327
Fri 01 Nov 2019 08:30:10 AM GMT -- oldpids: 20369 2348
Fri 01 Nov 2019 08:30:15 AM GMT -- oldpids: 20369 2373
Fri 01 Nov 2019 08:30:20 AM GMT -- oldpids: 20369 2394
Fri 01 Nov 2019 08:30:25 AM GMT -- oldpids: 20369 2415
```

Let's adjust the tunnel timeout in haproxy.cfg to 30s:

    tunnel timeout 30s

and connect with a different client and enter 'hello':

```console
$ websocat -E -n -t ws://127.0.0.1:4242/echo 
hello
handled by "127.0.0.1:9000"; you said "hello\n"
```

```console
Fri 01 Nov 2019 08:35:39 AM GMT -- oldpids: 20369 4507
Fri 01 Nov 2019 08:35:44 AM GMT -- oldpids: 20369 4541
Fri 01 Nov 2019 08:35:49 AM GMT -- oldpids: 20369 4563
Fri 01 Nov 2019 08:35:54 AM GMT -- oldpids: 20369 4581
Fri 01 Nov 2019 08:35:59 AM GMT -- oldpids: 20369 4631 4581
Fri 01 Nov 2019 08:36:04 AM GMT -- oldpids: 20369 4654 4581
Fri 01 Nov 2019 08:36:09 AM GMT -- oldpids: 20369 4680 4581
Fri 01 Nov 2019 08:36:14 AM GMT -- oldpids: 20369 4714 4581
Fri 01 Nov 2019 08:36:19 AM GMT -- oldpids: 20369 4733 4581
Fri 01 Nov 2019 08:36:24 AM GMT -- oldpids: 20369 4755
Fri 01 Nov 2019 08:36:29 AM GMT -- oldpids: 20369 4795
```

We are continously reloading haproxy. Pid `4581` is our open
connection to the websocket endpoint but in this case the process
disappears after ~30s as there has been no traffic for more than 30s -
we typed a single "hello".

## Memory Leaks?

I wanted to test for memory leaks so I created a malloc/calloc
interposer [library](alloc.c) that ballons malloc (or calloc)
requests. The goal here is to make leaking obvious. If you malloc 1MB,
then immediately free it all well and good. But if the leaks are
measured in small numbers of bytes it can take too long to see the net
effect. Ballooning everything by 1MB (configurable) allows any leak to
be immediately magnified.

### Build the interposer library

    $ make
	
### Change then script to LD_PRELOAD the library

```diff
$ git diff reload-proxy
diff --git a/BZ1743291/reload-proxy b/BZ1743291/reload-proxy
index ce66224..5be2c70 100755
--- a/BZ1743291/reload-proxy
+++ b/BZ1743291/reload-proxy
@@ -23,12 +23,10 @@ haproxy_binary=~/haproxy-1.8/haproxy
 
 reload_status=0
 if [ -n "$old_pids" ]; then
-  #LD_PRELOAD=$TOPDIR/liballoc.so ~/haproxy-1.8/haproxy -f $config_file -p $pid_file -x /var/tmp/haproxy/run/haproxy.sock -sf $old_pids
-  $haproxy_binary -f $config_file -p $pid_file -x /var/tmp/haproxy/run/haproxy.sock -sf $old_pids
+  LD_PRELOAD=$TOPDIR/liballoc.so $haproxy_binary -f $config_file -p $pid_file -x /var/tmp/haproxy/run/haproxy.sock -sf $old_pids
   reload_status=$?
 else
-  #LD_PRELOAD=$TOPDIR/liballoc.so ~/haproxy-1.8/haproxy -f $config_file -p $pid_file
-  $haproxy_binary -f $config_file -p $pid_file
+  LD_PRELOAD=$TOPDIR/liballoc.so $haproxy_binary -f $config_file -p $pid_file
   reload_status=$?
 fi
```
