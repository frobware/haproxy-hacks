I tested this new build but directly on my desktop, not a cluster. I
was able to reproduce the SIGABRT on my desktop. After switching to
the new build I was not able to generate a SIGABRT.

My scripts are captured here:

  https://github.com/frobware/haproxy-hacks/tree/master/BZ1810573

These are my reproducer steps (assuming you take the scripts from my repo):

# Start our haproxy backends
$ server node.js &

# Start haproxy
$ ./reload-haproxy

# Verify that haproxy is running
$ pgrep haproxy

# I am using hey to generate load 
$ hey -c 1000 -m GET -z 10h http://localhost:4242

# Repeatedly reload haproxy in a 1 while loop:
$ ./while-1-reload

Now we look for coredumps from haproxy:

$ while :; do coredumpctl list | cat | grep -w 6; sleep 2; date; done

Thu 19 Mar 2020 04:40:10 PM GMT
Thu 2020-03-19 15:27:19 GMT   25111  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 15:29:11 GMT   12877  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13511  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13535  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13547  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13559  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13645  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:36:44 GMT   13633  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy

I let this run and run accumulating SIGABRTs:

Thu 2020-03-19 16:45:02 GMT    7019  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    6983  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    7001  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    7097  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:46:12 GMT   28840  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy

At 16:47 I then switched to the new binary. To do this change the
reload-script - swap out the line in the reload-script that has:

  : ${HAPROXY:=$topdir/haproxy18-1.8.17-3.el7.x86_64/haproxy}

for

  : ${HAPROXY:=$topdir/haproxy18-1.8.17-4.el7.x86_64/haproxy}

I then let this run for 1 hour and did not see haproxy core dump with
signal 6 (ABRT) until "Thu 2020-03-19 17:50:17". This was the time I
switched back to the old binary ("haproxy18-1.8.17-3").

Thu 2020-03-19 16:45:02 GMT    7019  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    6983  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    7001  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:45:03 GMT    7097  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 16:46:12 GMT   28840  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:50:17 GMT   19924  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy

And letting this run and run with haproxy18-1.8.17-3 I see:

Thu 19 Mar 2020 06:09:58 PM GMT
Thu 2020-03-19 17:50:17 GMT   19924  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:51:13 GMT    6218  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:51:36 GMT   13489  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:51:37 GMT   13453  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:51:50 GMT   17775  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:51:50 GMT   17797  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:52:54 GMT    6079  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:53:21 GMT   14453  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:54:04 GMT   27993  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:54:04 GMT   28015  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:54:06 GMT   28518  1000  1000   6 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haprox

Looking at the last SIGABRT in gdb I see:

$ coredumpctl gdb 29965
           PID: 29965 (haproxy)
           UID: 1000 (aim)
           GID: 1000 (aim)
        Signal: 6 (ABRT)
     Timestamp: Thu 2020-03-19 18:01:02 GMT (9min ago)
  Command Line: /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy -f /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy.cfg -p /var/tmp/haproxy/run/haproxy.pid -x /var/tmp/haproxy/run/haproxy.sock -sf 29945 29916 29898 29880 29862 29844 29826 29808 29790 29772 29754 29735 15785
    Executable: /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
 Control Group: /user.slice/user-1000.slice/user@1000.service/gnome-terminal-server.service
          Unit: user@1000.service
     User Unit: gnome-terminal-server.service
         Slice: user-1000.slice
     Owner UID: 1000 (aim)
       Boot ID: 9cdcee89fdea41ad960e682904ffef7d
    Machine ID: 3ff4ac170c9447d4a10d272740ea39c8
      Hostname: spicy
       Storage: /var/lib/systemd/coredump/core.haproxy.1000.9cdcee89fdea41ad960e682904ffef7d.29965.1584640862000000.lz4
       Message: Process 29965 (haproxy) of user 1000 dumped core.
                
                Stack trace of thread 29969:
                #0  0x00007fd17b101e35 __GI_raise (libc.so.6)
                #1  0x00007fd17b0ec895 __GI_abort (libc.so.6)
                #2  0x00007fd17b14508f __libc_message (libc.so.6)
                #3  0x00007fd17b14c40c malloc_printerr (libc.so.6)
                #4  0x00007fd17b14de74 _int_free (libc.so.6)
                #5  0x00007fd17b1510bb tcache_thread_shutdown (libc.so.6)
                #6  0x00007fd17b69a4e6 start_thread (libpthread.so.0)
                #7  0x00007fd17b1c6163 __clone (libc.so.6)
                
                Stack trace of thread 29975:
                #0  0x00007fd17b6a374d __lll_lock_wait (libpthread.so.0)
                #1  0x00007fd17b69cdc4 __GI___pthread_mutex_lock (libpthread.so.0)
                #2  0x00007fd17b74e7ca _dl_open (ld-linux-x86-64.so.2)
                #3  0x00007fd17b2002e1 do_dlopen (libc.so.6)
                #4  0x00007fd17b200e09 __GI__dl_catch_exception (libc.so.6)
                #5  0x00007fd17b200ea3 __GI__dl_catch_error (libc.so.6)
                #6  0x00007fd17b2003e7 dlerror_run (libc.so.6)
                #7  0x00007fd17b20047a __GI___libc_dlopen_mode (libc.so.6)
                #8  0x00007fd17b6a524b pthread_cancel_init (libpthread.so.0)
                #9  0x00007fd17b6a5464 _Unwind_ForcedUnwind (libpthread.so.0)
                #10 0x00007fd17b6a3566 __GI___pthread_unwind (libpthread.so.0)
                #11 0x00007fd17b69b7b2 __do_cancel (libpthread.so.0)
                #12 0x000055be4631e485 n/a (/home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy)

GNU gdb (GDB) Fedora 8.3-7.fc30
Copyright (C) 2019 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-redhat-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy...
Missing separate debuginfo for /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Try: dnf --enablerepo='*debug*' install /usr/lib/debug/.build-id/d0/de6c87b2ee9a0f0afc9b1ad1283b64c73793f9.debug
Reading symbols from .gnu_debugdata for /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy...
(No debugging symbols found in .gnu_debugdata for /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy)
[New LWP 29969]
[New LWP 29975]
[New LWP 29970]
[New LWP 29974]
[New LWP 29972]
[New LWP 29977]
[New LWP 29965]
[New LWP 29976]
[New LWP 29968]

[Thread debugging using libthread_db enabled]
Using host libthread_db library "/lib64/libthread_db.so.1".

Core was generated by `/home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.'.
Program terminated with signal SIGABRT, Aborted.
#0  __GI_raise (sig=sig@entry=6) at ../sysdeps/unix/sysv/linux/raise.c:50
50	  return ret;
[Current thread is 1 (Thread 0x7fd179bd9700 (LWP 29969))]
(gdb) bt
(gdb) bt
#0  __GI_raise (sig=sig@entry=6) at ../sysdeps/unix/sysv/linux/raise.c:50
#1  0x00007fd17b0ec895 in __GI_abort () at abort.c:79
#2  0x00007fd17b14508f in __libc_message (action=action@entry=do_abort, fmt=fmt@entry=0x7fd17b253a7e "%s\n") at ../sysdeps/posix/libc_fatal.c:181
#3  0x00007fd17b14c40c in malloc_printerr (str=str@entry=0x7fd17b255728 "double free or corruption (fasttop)") at malloc.c:5366
#4  0x00007fd17b14de74 in _int_free (av=0x7fd17b28aba0 <main_arena>, p=0x55be4683fb10, have_lock=<optimized out>) at malloc.c:4278
#5  0x00007fd17b1510bb in tcache_thread_shutdown () at malloc.c:2978
#6  __malloc_arena_thread_freeres () at arena.c:952
#7  0x00007fd17b1543e0 in __libc_thread_freeres () at thread-freeres.c:38
#8  0x00007fd17b69a4e6 in start_thread (arg=<optimized out>) at pthread_create.c:493
#9  0x00007fd17b1c6163 in clone () at ../sysdeps/unix/sysv/

This comes from tcache_thread_shutdown(), not deinit_log_buffers() as
mentioned in comment #12. But comment #9 is a SIGABRT with a different
trace. So, although the stack trace is not identical to the patch, I'm
seeing that the new build fixes the SIGABRT.

Conclusion:

-- I ran haproxy18-1.8.17-3 for ~45 mins and saw many SIGABRTs.

-- I ran haproxy18-1.8.17-4 for ~45 mins and saw ZERO SIGABRTs.

-- I ran haproxy18-1.8.17-3 again and immediately saw many SIGABRTs.

Other observations:

I'm driving reload way harder/faster than we would normally do and I
also see quite a lot of SEGVs in either version (-3, or -4):

Thu 2020-03-19 17:56:29 GMT   10738  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:56:34 GMT   11719  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:56:37 GMT   12862  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:56:40 GMT   13620  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:56:45 GMT   15325  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:57:05 GMT   21969  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:57:37 GMT   31938  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:58:35 GMT   18224  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:58:35 GMT   18205  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:58:36 GMT   18537  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:58:38 GMT   19113  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:59:08 GMT   28852  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 17:59:28 GMT    2588  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 18:00:13 GMT   14684  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy

Switching to haproxy18-1.8.17-4 I also see SEGVs:

Thu 2020-03-19 18:01:05 GMT   30665  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.x86_64/haproxy
Thu 2020-03-19 18:37:20 GMT    9801  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:27 GMT   11912  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:37 GMT   15092  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:38 GMT   15294  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:48 GMT   18500  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:50 GMT   18960  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy
Thu 2020-03-19 18:37:51 GMT   19480  1000  1000  11 present   /home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.x86_64/haproxy

I don't want to muddy the waters though because it seems to me that
haproxy18-1.8.17-4 fixes the SIGABRTs.

For completeness here's a SEGv backtrace:

$ coredumpctl gdb 30665
Core was generated by `/home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-3.el7.'.
Program terminated with signal SIGSEGV, Segmentation fault.
#0  0x000056273701bc33 in listener_accept ()
[Current thread is 1 (Thread 0x7f20c666f700 (LWP 30672))]
(gdb) bt
#0  0x000056273701bc33 in listener_accept ()
#1  0x000056273703e10f in fd_process_cached_events ()
#2  0x0000562736fea2f2 in run_thread_poll_loop ()
#3  0x00007f20ca1344c0 in start_thread (arg=<optimized out>) at pthread_create.c:479
#4  0x00007f20c9c60163 in clone () at ../sysdeps/unix/sysv/linux/x86_64/clone.S:95

$ coredumpctl gdb 19480
Core was generated by `/home/aim/repos/github/frobware/haproxy-hacks/BZ1810573/haproxy18-1.8.17-4.el7.'.
Program terminated with signal SIGSEGV, Segmentation fault.
#0  0x00005557c4edecb3 in listener_accept ()
[Current thread is 1 (Thread 0x7fb4add98700 (LWP 19490))]
(gdb) bt
#0  0x00005557c4edecb3 in listener_accept ()
#1  0x00005557c4f0118f in fd_process_cached_events ()
#2  0x00005557c4ead342 in run_thread_poll_loop ()
#3  0x00007fb4b30604c0 in start_thread (arg=<optimized out>) at pthread_create.c:479
#4  0x00007fb4b2b8c163 in clone () at ../sysdeps/unix/sysv/linux/x86_64/clone.S:95

Is the SIGABRT the only core dump the customer is seeing?

(+ 16750 3000 1500)
