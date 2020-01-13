package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

var reapedPids = make([]int, 512)
var reapedPidCount = 0
var reapTime = 5 * time.Second

func main() {
	startTime := time.Now()

	defer func() {
		fmt.Fprintf(os.Stderr, "ran for %v seconds\n", time.Now().Sub(startTime).Seconds())
	}()

	go func() {
		var lastCount int = 0
		for {
			fmt.Printf("%s: reaped pids/%vs: %d\n", time.Now(), reapTime.Seconds(), reapedPidCount-lastCount)
			lastCount = reapedPidCount
			time.Sleep(reapTime)
		}
	}()

	go OpenShiftStartReaper()

	for {
		cmd := exec.Command("/bin/reload-haproxy")

		if _, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "cmd.CombinedOutput() pid == %v\n", cmd.Process.Pid)
			sort.Ints(reapedPids)
			fmt.Fprintf(os.Stderr, "recenty reaped pids (sorted): %+v\n", reapedPids)
			panic(err.Error())
		} else {
			//fmt.Printf("cmd with pid %v completed\n%s\n", cmd.Process.Pid, out)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func StartReaper() {
	if os.Getpid() != 1 {
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGCHLD)

	for {
		<-sigs

		for {
			time.Sleep(reapTime)
			cpid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)

			for err == syscall.EINTR {
				cpid, err = syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
			}

			if cpid < 1 {
				break
			}

			reapedPids[reapedPidCount%len(reapedPids)] = cpid
			reapedPidCount += 1
			//fmt.Printf("StartReaper: reaped child %v\n", cpid)
		}
	}
}

func OpenShiftStartReaper() {
	if os.Getpid() == 1 {
		//klog.V(4).Infof("Launching reaper")
		go func() {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGCHLD)
			for {
				// Wait for a child to terminate
				<-sigs
				time.Sleep(reapTime)
				//klog.V(4).Infof("Signal received: %v", sig)
				for {
					// Reap processes
					cpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
					if cpid < 1 {
						break
					}

					reapedPids[reapedPidCount%len(reapedPids)] = cpid
					reapedPidCount += 1
					//fmt.Printf("Reaped process with pid %d\n", cpid)
				}
			}
		}()
	}
}
