package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	lifetime        = flag.Duration("lifetime", 0, "Lifetime for a connection (0 == indefinite)")
	interval        = flag.Duration("interval", 3*time.Second, "Interval between successive connections")
	exitOnConnError = flag.Bool("exitOnConnError", false, "Exit on connection error")
)

type request struct {
	Host              string
	WebSocketLifetime time.Duration
}

func startWSPingPong(r request) {
	if r.WebSocketLifetime == 0 {
		r.WebSocketLifetime = time.Hour
	}

	u := url.URL{Scheme: "wss", Host: r.Host, Path: "/ws"}

	log.Printf("connecting to %s, lifetime: %s\n", u.String(), r.WebSocketLifetime.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		switch *exitOnConnError {
		case true:
			log.Fatalf("dial: %q: %v\n", u.String(), err)
		default:
			log.Printf("dial: %q: %v\n", u.String(), err)
			return
		}
	}

	go func() {
		var firstMsgRecv bool

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("read error: %v\n", err)
				return
			}
			if !firstMsgRecv {
				log.Printf("recv: %q\n", message)
				firstMsgRecv = true
			}
		}
	}()

	ticker1s := time.NewTicker(time.Second)
	tickerNs := time.NewTicker(r.WebSocketLifetime)

	defer c.Close()

	for {
		select {
		case t := <-ticker1s.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Printf("write error: %v", err)
				return
			}
		case <-tickerNs.C:
			return
		}
	}
}

var addr = flag.String("addr", "route1.int.frobware.com:8443", "service address")

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	go func() {
		select {
		case <-interrupt:
			os.Exit(1)
		}
	}()

	for {
		go func() {
			startWSPingPong(request{
				Host:              *addr,
				WebSocketLifetime: *lifetime,
			})
		}()

		time.Sleep(*interval)
	}
}

// pod:container_memory_usage_bytes:sum{namespace="openshift-ingress"}
