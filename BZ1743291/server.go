// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:9000", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func home(w http.ResponseWriter, r *http.Request) {
	log.Printf("home %q", r.RemoteAddr)
	w.WriteHeader(200)
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("remote: %q, read error: %v", c.RemoteAddr(), err)
			break
		}
		log.Printf("recv message from %q\n", c.RemoteAddr())
		err = c.WriteMessage(mt, []byte(fmt.Sprintf("handled by %q; you said %q", c.LocalAddr(), message)))
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Printf("remote %q: write error: %v", c.RemoteAddr(), err)
			break
		}
	}

	log.Printf("remote %q disappeared", c.RemoteAddr())
}

func main() {
	flag.Parse()
	log.SetFlags(0x7f)
	http.HandleFunc("/", home)
	http.HandleFunc("/echo", echo)
	go func() {
		log.Fatal(http.ListenAndServe("127.0.0.1:9000", nil))
	}()
	go func() {
		log.Fatal(http.ListenAndServe("127.0.0.1:9001", nil))
	}()

	log.Fatal(http.ListenAndServe("127.0.0.1:9002", nil))
}
