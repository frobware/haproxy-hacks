# haproxy-hacks
Dumping ground for haproxy experiments

# Process table before adding new frontends

```console
$ docker exec -it epic_torvalds bash
[root@ba6aad638d2b /]# ps -ax
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d
    8 ?        Sl     0:00 /usr/local/bin/dataplaneapi --host 0.0.0.0 --port 5555 --haproxy-bin /usr/sbin/haproxy --config-file /etc/haproxy/haproxy.cfg --reload-cmd kill -SI
    9 ?        Sl     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d
   29 ?        Ss     0:00 bash
   43 ?        R+     0:00 ps -ax
```

# Process table after adding new frontends

```console
root@30d75d388799 /]# ps -ax
  PID TTY      STAT   TIME COMMAND
    1 ?        Ss     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d -sf 8 9 -x /var/lib/haproxy/run/haproxy.sock
    8 ?        Sl     0:00 /usr/local/bin/dataplaneapi --host 0.0.0.0 --port 5555 --haproxy-bin /usr/sbin/haproxy --config-file /etc/haproxy/haproxy.cfg --reload-cmd kill -SI
   42 ?        Sl     0:00 /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg -d -sf 8 9 -x /var/lib/haproxy/run/haproxy.sock
   50 ?        Ss     0:00 bash
   64 ?        R+     0:00 ps -ax
```

So, new pids.
