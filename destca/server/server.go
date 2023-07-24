package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
)

//go:embed *.html
var BackendFS embed.FS

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

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,DELETE,PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
}

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		fs.ServeHTTP(w, r)
	}
}

func writeResponse(w http.ResponseWriter, req *http.Request) {
	if false {
		fmt.Printf("Headers: ")
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Printf("%v: %v\n", name, h)
			}
		}
		fmt.Println()
	}
	fmt.Fprintf(w, "%s %s %s://%s:%s%s\n",
		req.Proto,
		req.RemoteAddr,
		req.Header.Get("X-Forwarded-Proto"),
		req.Header.Get("X-Forwarded-Host"),
		req.Header.Get("X-Forwarded-Port"),
		req.URL)
}

func main() {
	crtFile := lookupEnv("TLS_CRT", defaultTLSCrt)
	keyFile := lookupEnv("TLS_KEY", defaultTLSKey)

	http.Handle("/", cors(http.FileServer(http.FS(BackendFS))))

	http.HandleFunc("/healthy", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, req *http.Request) {
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		enableCors(w)
		writeResponse(w, req)
		log.Println(req.Proto, req.URL, "connection from", req.RemoteAddr)
	})

	log.Println("GODEBUG", lookupEnv("GODEBUG", "<unset>"))

	go func() {
		port := lookupEnv("HTTP_PORT", defaultHTTPPort)
		log.Printf("Listening on port %v\n", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		port := lookupEnv("HTTPS_PORT", defaultHTTPSPort)
		log.Printf("Listening securely on port %v\n", port)

		if err := http.ListenAndServeTLS(":"+port, crtFile, keyFile, nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}
