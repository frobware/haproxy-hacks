package main_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

var permittedHeaderValueTemplate = `^(?:%(?:%|(?:\{[-+]?[QXE](?:,[-+]?[QXE])*\})?\[(?:XYZ\.hdr\([0-9A-Za-z-]+\)|ssl_c_der)(?:,(?:lower|base64))*\])|[^%[:cntrl:]])+$`
var permittedRequestHeaderValueRE = regexp.MustCompile(strings.Replace(permittedHeaderValueTemplate, "XYZ", "req", 1))
var permittedResponseHeaderValueRE = regexp.MustCompile(strings.Replace(permittedHeaderValueTemplate, "XYZ", "res", 1))

func validateInput(input string) bool {
	re := regexp.MustCompile(`^(?:%(?:%|(?:\{[-+]?[QXE](?:,[-+]?[QXE])*\})?\[(?:XYZ\.hdr\([0-9A-Za-z-]+\)|ssl_c_der)(?:,(?:lower|base64))*\])|[^%[:cntrl:]])+$`)
	return re.MatchString(input)
}

func TestHeaderValues(t *testing.T) {
	type HeaderValueTest struct {
		description string
		isValid     bool
		input       string
	}

	quoteValue := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}

	tests := []HeaderValueTest{
		{
			input:   "%%[ssl_c_der,lower]",
			isValid: true,
		},
		{description: "empty value", input: ``, isValid: false},
		{description: "single character", input: `a`, isValid: true},
		{description: "multiple characters", input: `abc`, isValid: true},
		{description: "multiple words without escaped space", input: `abc def`, isValid: true},
		{description: "multiple words with escaped space", input: `abc\ def`, isValid: true},
		{description: "multiple words each word quoted", input: `"abc"\ "def"`, isValid: true},
		{description: "multiple words each word quoted and with an embedded space", input: `"abc "\ "def "`, isValid: true},
		{description: "multiple words each word one double quoted and other single quoted and with an embedded space", input: `"abc "\ 'def '`, isValid: true},
		{description: "single % character", input: `%`, isValid: false},
		{description: "escaped % character", input: `%%`, isValid: true},
		{description: "escaped % and only a % character", input: `%%%`, isValid: false},
		{description: "two literal % characters", input: `%%%%`, isValid: true},
		{description: "zero percent", input: `%%%%%%0`, isValid: true},
		{description: "escaped expression", input: `%%[XYZ.hdr(Host)]\ %[XYZ.hdr(Host)]`, isValid: true},
		{description: "simple empty expression", input: `%[]`, isValid: false},
		{description: "nested empty expressions", input: `%[%[]]`, isValid: false},
		{description: "empty quoted value", input: `%{+Q}`, isValid: false},
		{description: "quoted value", input: `%{+Q}foo`, isValid: false},
		{description: "hdr with empty field", input: `%[XYZ.hdr()]`, isValid: false},
		{description: "hdr with percent field", input: `%[XYZ.hdr(%)]`, isValid: false},
		{description: "hdr with known field", input: `%[XYZ.hdr(Host)]`, isValid: true},
		{description: "hdr with syntax error", input: `%[XYZ.hdr(Host]`, isValid: false},
		{description: "hdr with url", input: `%[XYZ.hdr(Host)] http://url/hack`, isValid: true},

		{description: "hdr with valid X-XSS-Protection value", input: `1;mode=block`, isValid: true},
		{description: "hdr with valid Content-Type value", input: `text/plain,text/html`, isValid: true},
		{description: "hdr with url", input: `text/plain,text/html`, isValid: true},

		{description: "incomplete expression", input: `%[req`, isValid: false},
		{description: "quoted field", input: `%[XYZ.hdr(%{+Q}Host)]`, isValid: false},
		{description: "value with conditional expression", input: `%[XYZ.hdr(Host)] if foo`, isValid: true},
		{description: "value with what looks like a conditional expression", input: `%[XYZ.hdr(Host)]\ if\ foo`, isValid: true},
		{description: "unsupported fetcher and converter", input: `%[date(3600),http_date]`, isValid: false},
		{description: "not allowed sample fetches", input: `%[foo,lower]`, isValid: false},
		{description: "not allowed converters", input: `%[req.hdr(host),foo]`, isValid: false},
		{description: "missing parentheses or braces", input: `%{Q[req.hdr(host)]`, isValid: false},
		{description: "missing parentheses or braces", input: `%Q}[req.hdr(host)]`, isValid: false},
		{description: "missing parentheses or braces", input: `%{{Q}[req.hdr(host)]`, isValid: false},
		{description: "missing parentheses or braces", input: `%[req.hdr(host)`, isValid: false},
		{description: "missing parentheses or braces", input: `%req.hdr(host)]`, isValid: false},
		{description: "missing parentheses or braces", input: `%[req.hdrhost)]`, isValid: false},
		{description: "missing parentheses or braces", input: `%[req.hdr(host]`, isValid: false},
		{description: "missing parentheses or braces", input: `%[req.hdr(host`, isValid: false},
		{description: "missing parentheses or braces", input: `%{req.hdr(host)}`, isValid: false},
		{description: "parameters for a sample fetch that doesn't take parameters", input: `%[ssl_c_der(host)]`, isValid: false},
		{description: "dangerous sample fetchers and converters", input: `%[env(FOO)]`, isValid: false},
		{description: "dangerous sample fetchers and converters", input: `%[req.hdr(host),debug()]`, isValid: false},
		{description: "extra comma", input: `%[req.hdr(host),,lower]`, isValid: false},

		// CR and LF are not allowed in header value as per RFC https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.4
		// {description: "carriage return", input: "\r", isValid: false},
		// {description: "CRLF", input: "\r\n", isValid: false},

		// This value is allowed in haproxy.config and does not cause haproxy to crash. The input strings are sanitized by escaping single quotes in the value and then
		// single quoting the whole value which helps haproxy not to crash in the event of an invalid string provided.
		{description: "environment variable with a bracket missing", input: "${NET_COOLOCP_HOSTPRIMARY", isValid: true},
		{description: "value with conditional expression and env var", input: `%[XYZ.hdr(Host)] if ${NET_COOLOCP_HOSTPRIMARY`, isValid: true},
		{description: "value with what looks like a conditional expression and env var", input: `%[XYZ.hdr(Host)]\ if\ ${NET_COOLOCP_HOSTPRIMARY}`, isValid: true},

		{description: "sample value", input: "%ci:%cp [%tr] %ft %ac/%fc %[fc_err]/\\%[ssl_fc_err,hex]/%[ssl_c_err]/%[ssl_c_ca_err]/%[ssl_fc_is_resumed] \\%[ssl_fc_sni]/%sslv/%sslc", isValid: false},
		{description: "interpolation of T i.e %T", input: `%T`, isValid: false},

		// url
		// regex does not check validity of url in a header value.
		{description: "hdr with url", input: `%[XYZ.hdr(Host)] http:??//url/hack`, isValid: true},
		{description: "hdr with url", input: `http:??//url/hack`, isValid: true},

		// spaces and tab
		// regex allows spaces before and after. The reason is that after a dynamic value is provided someone might provide a condition `%[XYZ.hdr(Host)] if foo` which would have
		// spaces after the dynamic value and if condition.
		// tab is rejected as control characters are not allowed by the regex.
		{description: "space before and after the value", input: ` T `, isValid: true},
		{description: "double space before and after the value", input: `  T  `, isValid: true},
		{description: "tab before and after the value", input: `	T	`, isValid: false},
		{description: "tab before and after the value", input: "\tT\t", isValid: false},
	}

	var requestTypes = []struct {
		description          string
		regexp               *regexp.Regexp
		testInputSubstituter func(s string) string
	}{{
		description:          "request",
		regexp:               permittedRequestHeaderValueRE,
		testInputSubstituter: func(s string) string { return strings.ReplaceAll(s, "XYZ", "req") },
	}, {
		description:          "response",
		regexp:               permittedResponseHeaderValueRE,
		testInputSubstituter: func(s string) string { return strings.ReplaceAll(s, "XYZ", "res") },
	}}

	var haproxyContent []string // local testing only

	for _, rt := range requestTypes {
		t.Run(rt.description, func(t *testing.T) {
			for j, tc := range tests {
				t.Run(tc.description, func(t *testing.T) {
					input := rt.testInputSubstituter(tc.input)
					if got := rt.regexp.MatchString(input); got != tc.isValid {
						t.Errorf("%q: expected %v, got %t", input, tc.isValid, got)
					}
					haproxyTestIdent := fmt.Sprintf("# %v %s", j, t.Name())
					haproxyTestItem := fmt.Sprintf("http-%s set-header Test%shdr_%v %s", rt.description, rt.description, j, quoteValue(input))
					if !tc.isValid {
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
