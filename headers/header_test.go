package main_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

var permittedHeaderValueTemplate = `^(?:%(?:%|(?:\{[-+]?[QXE](,[-+]?[QXE])*\})?\[(?:XYZ\.hdr\([0-9A-Za-z-]+\)|ssl_c_der)(?:,(?:lower|base64))*\])|[^%[:cntrl:]])*$`
var permittedRequestHeaderValueRE = regexp.MustCompile(strings.Replace(permittedHeaderValueTemplate, "XYZ", "req", 1))
var permittedResponseHeaderValueRE = regexp.MustCompile(strings.Replace(permittedHeaderValueTemplate, "XYZ", "res", 1))

func TestHeaderValues(t *testing.T) {
	type HeaderValueTest struct {
		description string
		validInput  bool
		input       string
	}

	quoteValue := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}

	tests := []HeaderValueTest{
		// {description: "empty string", input: ``, valid: true},
		{description: "single character", input: `a`, validInput: true},
		{description: "multiple characters", input: `abc`, validInput: true},
		{description: "multiple words without escaped space", input: `abc def`, validInput: true},
		{description: "multiple words with escaped space", input: `abc\ def`, validInput: true},
		{description: "multiple words each word quoted", input: `"abc"\ "def"`, validInput: true},
		{description: "multiple words each word quoted and with an embedded space", input: `"abc "\ "def "`, validInput: true},
		{description: "single % character", input: `%`, validInput: false},
		{description: "escaped % character", input: `%%`, validInput: true}, // hmm?
		{description: "escaped % and only a % character", input: `%%%`, validInput: false},
		{description: "two literal % characters", input: `%%%%`, validInput: true},
		{description: "zero percent", input: `%%%%%%0`, validInput: true},
		{description: "escaped expression", input: `%%[XYZ.hdr(Host)]\ %[XYZ.hdr(Host)]`, validInput: true},
		{description: "simple empty expression", input: `%[]`, validInput: false},
		{description: "nested empty expressions", input: `%[%[]]`, validInput: false},
		{description: "empty quoted value", input: `%{+Q}`, validInput: false},
		{description: "quoted value", input: `%{+Q}foo`, validInput: false},
		{description: "hdr with empty field", input: `%[XYZ.hdr()]`, validInput: false},
		{description: "hdr with percent field", input: `%[XYZ.hdr(%)]`, validInput: false},
		{description: "hdr with known field", input: `%[XYZ.hdr(Host)]`, validInput: true},
		{description: "hdr with syntax error", input: `%[XYZ.hdr(Host]`, validInput: false},
		{description: "incomplete expression", input: `%[req`, validInput: false},
		{description: "quoted field", input: `%[XYZ.hdr(%{+Q}Host)]`, validInput: false},
		{description: "value with conditional expression", input: `%[XYZ.hdr(Host)] if foo`, validInput: true},
		{description: "value with what looks like a conditional expression", input: `%[XYZ.hdr(Host)]\ if\ foo`, validInput: true},
		{description: "unsuported fetcher", input: `%[date(3600),http_date]`, validInput: false},
	}

	var requests = []struct {
		name             string
		regexp           *regexp.Regexp
		inputSubstituter func(s string) string
	}{{
		name:             "request",
		regexp:           permittedRequestHeaderValueRE,
		inputSubstituter: func(s string) string { return strings.ReplaceAll(s, "XYZ", "req") },
	}, {
		name:             "response",
		regexp:           permittedResponseHeaderValueRE,
		inputSubstituter: func(s string) string { return strings.ReplaceAll(s, "XYZ", "res") },
	}}

	var haproxyContent []string // local testing only

	for _, rt := range requests {
		t.Run(rt.name, func(t *testing.T) {
			for j, tc := range tests {
				t.Run(tc.description, func(t *testing.T) {
					input := rt.inputSubstituter(tc.input)
					if got := rt.regexp.MatchString(input); got != tc.validInput {
						t.Errorf("%q: expected %v, got %t", input, tc.validInput, got)
					}
					haproxyTestIdent := fmt.Sprintf("# %v %s", j, t.Name())
					haproxyTestItem := fmt.Sprintf("http-%s set-header Test%shdr_%v %s", rt.name, rt.name, j, quoteValue(input))
					if !tc.validInput {
						haproxyTestItem = "# " + haproxyTestItem
					}
					haproxyContent = append(haproxyContent, haproxyTestIdent, haproxyTestItem, "\n")
				})
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

	fmt.Fprintf(bfd, `
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
%s
`[1:], strings.Join(haproxyContent, "\n"))
}
