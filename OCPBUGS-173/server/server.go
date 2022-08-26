// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	defaultHTTPPort  = "8080"
	defaultHTTPSPort = "8443"
	defaultTLSCrt    = "/etc/serving-cert/tls.crt"
	defaultTLSKey    = "/etc/serving-cert/tls.key"
)

func lookupEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.Header {
		log.Printf("[%s] %s: %s\n", r.RemoteAddr, http.CanonicalHeaderKey(k), v)
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		if strings.HasPrefix(string(message), "headers") {
			for k, v := range r.Header {
				err = c.WriteMessage(mt, []byte(fmt.Sprintf("[%s] %s: %s", r.RemoteAddr, http.CanonicalHeaderKey(k), v)))
				if err != nil {
					log.Println("write error:", err)
					break
				}
			}
		}
		err = c.WriteMessage(mt, []byte("echo: "+string(message)))
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.Header {
		log.Printf("[%s] %s: %s\n", r.RemoteAddr, http.CanonicalHeaderKey(k), v)
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		homeTemplate.Execute(w, "wss://"+r.Host+"/echo")
	} else {
		homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)

	go func() {
		port := lookupEnv("HTTP_PORT", defaultHTTPPort)
		log.Printf("Listening on port %v\n", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		crtFile := lookupEnv("TLS_CRT", defaultTLSCrt)
		keyFile := lookupEnv("TLS_KEY", defaultTLSKey)

		port := lookupEnv("HTTPS_PORT", defaultHTTPSPort)
		log.Printf("Listening securely on port %v\n", port)

		if err := http.ListenAndServeTLS(":"+port, crtFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
   <head>
      <meta charset="utf-8">
      <script>
	 window.addEventListener("load", function(evt) {
	     var output = document.getElementById("output");
	     var input = document.getElementById("input");
	     var ws;

	     var print = function(message) {
		var d = document.createElement("div");
		d.textContent = message;
		output.appendChild(d);
		output.scroll(0, output.scrollHeight);
	     };

	     document.getElementById("open").onclick = function(evt) {
		if (ws) {
		    return false;
		}
		ws = new WebSocket("{{.}}");
		ws.onopen = function(evt) {
		    print("OPEN");
		}
		ws.onclose = function(evt) {
		    print("CLOSE");
		    ws = null;
		}
		ws.onmessage = function(evt) {
		    print("RESPONSE: " + evt.data);
		}
		ws.onerror = function(evt) {
		    print("ERROR: " + evt.data);
		}
		return false;
	     };

	     document.getElementById("send").onclick = function(evt) {
		if (!ws) {
		    return false;
		}
		print("SEND: " + input.value);
		ws.send(input.value);
		return false;
	     };

	     document.getElementById("close").onclick = function(evt) {
		if (!ws) {
		    return false;
		}
		ws.close();
		return false;
	     };
	 });
      </script>
   </head>
   <body>
      <table>
	 <tr>
	    <td valign="top" width="50%">
	       <p>Click "Open" to create a connection to the server,
		  "Send" to send a message to the server and "Close" to close the connection.
		  You can change the message and send multiple times.
	       <p>
	       <form>
		  <button id="open">Open</button>
		  <button id="close">Close</button>
		  <p><input id="input" type="text" value="Hello world!">
		     <button id="send">Send</button>
	       </form>
	    </td>
	    <td valign="top" width="50%">
	       <div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
	    </td>
	 </tr>
      </table>
   </body>
</html>
`))
