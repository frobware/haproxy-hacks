package main_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

var headerRe = regexp.MustCompile(`^(?:%(?:%|(?:\{[-+]?[QXE](,[-+]?[QXE])*\})?\[(?:req\.hdr\([0-9A-Za-z-]+\)|ssl_c_der)(?:,(?:lower|base64))*\])|[^%[:cntrl:]])*$`)

func validateHeaderValue(input string) bool {
	return headerRe.MatchString(input)
}

func TestSamples(t *testing.T) {
	type HeaderValueTest struct {
		description string
		valid       bool
		input       string
	}

	tests := []HeaderValueTest{
		{description: "empty string", input: ``, valid: false},
		{description: "single character", input: `a`, valid: true},
		{description: "multiple characters", input: `abc`, valid: true},
		{description: "multiple words without escaped space", input: `abc def`, valid: false},
		{description: "multiple words with escaped space", input: `abc\ def`, valid: true},
		{description: "multiple words each word quoted", input: `"abc"\ "def"`, valid: true},
		{description: "multiple words each word quoted and with an embedded space", input: `"abc "\ "def "`, valid: true},
		{description: "unescaped % character", input: `%`, valid: false},
		{description: "escaped %% character", input: `%%`, valid: true}, // hmm?
		{description: "escaped % and only a % character", input: `%%%`, valid: false},
		{description: "two % characters", input: `%%%%`, valid: true},
		{description: "zero percent", input: `%%%%%%0`, valid: true},
		{description: "escaped expression", input: `%%[src]\ %[src]`, valid: true},
		{description: "simple empty expression", input: `%[]`, valid: false},
		{description: "nested empty expressions", input: `%[%[]]`, valid: false},
		{description: "empty quoted value", input: `%{+Q}`, valid: false},
		{description: "quoted value", input: `%{+Q}foo`, valid: false},
		{description: "request hdr with empty field", input: `%[req.hdr()]`, valid: false},
		{description: "request hdr with percent field", input: `%[req.hdr(%)]`, valid: true},
		{description: "request hdr with known field", input: `%[req.hdr(Host)]`, valid: true},
		{description: "request hdr with syntax error", input: `%[req.hdr(Host]`, valid: false},
		{description: "incomplete expression", input: `%[req`, valid: false},
		{description: "quoted field", input: `%[req.hdr(%{+Q}Host)]`, valid: true},
		{description: "value with conditional expression", input: `%[req.hdr(Host)] if foo`, valid: false},
		{description: "value with what looks like a conditional expression", input: `%[req.hdr(Host)]\ if\ foo`, valid: true},
		{description: "unsuported fetcher", input: `%[date(3600),http_date]`, valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			if got := validateHeaderValue(tc.input); got != tc.valid {
				t.Errorf("%q: expected %v, got %t", tc.input, tc.valid, got)
			}
		})
	}

	// ignore the remainder of this...
	// This generates haproxy.config for absolute verification
	// which you can use to test "for real".

	bfd, err := os.Create("haproxy.cfg")
	if err != nil {
		panic(err)
	}
	defer bfd.Close()

	fmt.Fprintln(bfd, `
global
  log /dev/stdout format raw local0 debug

defaults
  log global
  mode http
  option httplog
  option logasap
  timeout client 5s
  timeout connect 5s
  timeout server 5s

frontend public
  log global
  bind *:8080
  mode http
  tcp-request content accept if HTTP
  tcp-request inspect-delay 5s
  default_backend default

backend default
  log global
  mode http
  server s1 :9090

`[1:])

	for i, tc := range tests {
		fmt.Fprintf(bfd, "  # %q\n", tc.description)
		if tc.valid {
			fmt.Fprintf(bfd, "  http-request set-header Testhdr_%d %s\n\n", i, tc.input)
		} else {
			fmt.Fprintf(bfd, "  # http-request set-header Testhdr_%d %s\n\n", i, tc.input)
		}
	}
}
