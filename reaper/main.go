package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	startTime := time.Now()

	defer func() {
		fmt.Fprintf(os.Stderr, "ran for %v seconds\n", time.Now().Sub(startTime).Seconds())
	}()

	go StartReaper()

	for {
		cmd := exec.Command("/bin/reload-haproxy")

		if _, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "cmd.CombinedOutput() pid == %v\n", cmd.Process.Pid)
			panic(err.Error())
		} else {
			// fmt.Println(string(out))
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
			cpid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)

			for err == syscall.EINTR {
				cpid, err = syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
			}

			if cpid < 1 {
				break
			}

			fmt.Printf("StartReaper: reaped child %v\n", cpid)
		}
	}
}
