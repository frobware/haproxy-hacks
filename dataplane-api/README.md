Showcases `cmd.CombinedOutput()` failing.
    
The `StartReaper()` go routine occasionally reaps the PID that cmd is
waiting on:
    
```console
StartReaper: reaped child 1754
StartReaper: reaped child 1768
StartReaper: reaped child 1782
StartReaper: reaped child 1796
StartReaper: reaped child 1810
StartReaper: reaped child 1824
StartReaper: reaped child 1838
StartReaper: reaped child 1852
StartReaper: reaped child 1866
StartReaper: reaped child 1880
StartReaper: reaped child 2031
cmd.CombinedOutput() pid == 2031
ran for 5.143265703 seconds
panic: wait: no child processes

goroutine 1 [running]:
main.main()
	/home/aim/haproxy-hacks/reaper/main.go:26 +0x1f6
make: *** [Makefile:4: run] Error 2
```
