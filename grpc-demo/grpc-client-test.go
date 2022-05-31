package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	grpcinterop "github.com/frobware/haproxy-hacks/grpc-demo/grpc-interop"
)

var supportedTestNames = []string{
	"cancel_after_begin",
	"cancel_after_first_response",
	"client_streaming",
	"custom_metadata",
	"empty_unary",
	"large_unary",
	"ping_pong",
	"server_streaming",
	"special_status_message",
	"status_code_and_message",
	"timeout_on_sleeping_server",
	"unimplemented_method",
	"unimplemented_service",
}

var domain = flag.String("domain", "example.com", "ingresscontroller domain")
var timeout = flag.Duration("timeout", 5*time.Second, "connection timeout")

func main() {
	flag.Parse()

	for _, test := range []struct {
		routeType  string
		dialParams grpcinterop.DialParams
	}{{
		routeType: "edge",
		dialParams: grpcinterop.DialParams{
			Port:     443,
			UseTLS:   true,
			Insecure: true,
		},
	}, {
		routeType: "h2c",
		dialParams: grpcinterop.DialParams{
			Port:     80,
			UseTLS:   false,
			Insecure: true,
		},
	}, {
		routeType: "reencrypt",
		dialParams: grpcinterop.DialParams{
			Port:     443,
			UseTLS:   true,
			Insecure: true,
		},
	}, {
		routeType: "passthrough",
		dialParams: grpcinterop.DialParams{
			Port:     443,
			UseTLS:   true,
			Insecure: true,
		},
	}} {
		test.dialParams.Host = fmt.Sprintf("grpc-interop-%s.%s", test.routeType, *domain)
		log.Printf("%+v\n", test.dialParams)
		if err := execTests(test.dialParams, *timeout, supportedTestNames); err != nil {
			log.Fatalf("error host=%q, err=%v\n", test.dialParams.Host, err)
		}
	}
}

func execTests(dialParams grpcinterop.DialParams, timeout time.Duration, testNames []string) error {
	for _, name := range testNames {
		log.Printf("%s/%s\n", dialParams.Host, name)
		conn, err := grpcinterop.Dial(dialParams, timeout)
		if err != nil {
			return fmt.Errorf("error: connection failed: %v, retrying...", err)
		}
		defer conn.Close()
		if err := grpcinterop.ExecTestCase(conn, name); err != nil {
			return fmt.Errorf("gRPC interop test case %q via host %q failed: %v, retrying...", name, dialParams.Host, err)
		}
	}

	return nil
}
