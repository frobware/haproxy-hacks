package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// connection tracing and timing info lifted from httpstat:
//   https://github.com/davecheney/httpstat/blob/master/main.go

const (
	httpsTemplate = `` +
		`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
		`------------------------------------------------------------------------------------` + "\n" +
		`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
		`            |                |               |                   |                  |` + "\n" +
		`   namelookup:%s      |               |                   |                  |` + "\n" +
		`                       connect:%s     |                   |                  |` + "\n" +
		`                                   pretransfer:%s         |                  |` + "\n" +
		`                                                     starttransfer:%s        |` + "\n" +
		`                                                                                total:%s` + "\n"

	httpTemplate = `` +
		`   DNS Lookup   TCP Connection   Server Processing   Content Transfer` + "\n" +
		`[ %s  |     %s  |        %s  |       %s  ]` + "\n" +
		`             |                |                   |                  |` + "\n" +
		`    namelookup:%s      |                   |                  |` + "\n" +
		`                        connect:%s         |                  |` + "\n" +
		`                                      starttransfer:%s        |` + "\n" +
		`                                                                 total:%s` + "\n"
)

// Fetcher fetches HTML documents.
type Fetcher interface {
	Fetch(request) *result
}

// Ensure fetcher is a Fetcher.
var _ Fetcher = (*fetcher)(nil)

var (
	verbose = flag.Bool("v", false, "Verbose")
	workers = flag.Int("workers", 50, "number of workers")
	timeout = flag.Duration("timeout", 250*time.Millisecond, "client.GET timeout")
	queue   = flag.Int("queue", 100, "concurent queue depth")
	repeat  = flag.Bool("repeat", false, "Insert finished jobs back into the queue")
)

type fetcher struct {
	MaxConnectionsPerHost int
	FetchTimeout          time.Duration
	Fetcher
}

type request struct {
	Fetcher
	URL string
}

func fprintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}

var fmtaSpacer string = strings.Repeat(" ", 9)
var fmtbSpacer string = strings.Repeat(" ", 9)

func (r result) Print() {
	fmta := func(d time.Duration) string {
		return fmt.Sprintf("%7dms", int(d/time.Millisecond))
	}

	fmtb := func(d time.Duration) string {
		return fmt.Sprintf("%-9s", strconv.Itoa(int(d/time.Millisecond))+"ms")
	}

	var dnsLookup string = fmtaSpacer
	var tcpConnection string = fmtaSpacer
	var tlsHandshake string = fmtaSpacer
	var serverProcessing string = fmtaSpacer
	var contentTransfer string = fmtaSpacer
	var nameLookup string = fmtbSpacer
	var connect string = fmtbSpacer
	var preTransfer string = fmtbSpacer
	var startTransfer string = fmtbSpacer
	var total string = fmtbSpacer

	if !r.t1.IsZero() {
		dnsLookup = fmta(r.t1.Sub(r.t0))
	}
	if !r.t2.IsZero() {
		tcpConnection = fmta(r.t2.Sub(r.t1))
	}
	if !r.t6.IsZero() {
		tlsHandshake = fmta(r.t6.Sub(r.t5))
	}
	if !r.t4.IsZero() {
		serverProcessing = fmta(r.t4.Sub(r.t3))
	}
	if !r.t7.IsZero() {
		contentTransfer = fmta(r.t7.Sub(r.t4))
	}

	if !r.t1.IsZero() {
		nameLookup = fmtb(r.t1.Sub(r.t0))
	}
	if !r.t2.IsZero() {
		connect = fmtb(r.t2.Sub(r.t0))
	}
	if !r.t3.IsZero() {
		preTransfer = fmtb(r.t3.Sub(r.t0))
	}
	if !r.t4.IsZero() {
		startTransfer = fmtb(r.t4.Sub(r.t0))
	}
	if !r.t7.IsZero() {
		total = fmtb(r.t7.Sub(r.t0))
	}

	fprintf(httpsTemplate,
		dnsLookup,
		tcpConnection,
		tlsHandshake,
		serverProcessing,
		contentTransfer,
		nameLookup,
		connect,
		preTransfer,
		startTransfer,
		total,
	)

}

func (r result) TotalTime() time.Duration {
	return r.t7.Sub(r.t0)
}

type connectionTrace struct {
	DNSStart, DNSDone, ConnectStart, ConnectDone, GotConn, GotFirstResponseByte, TLSHandshakeStart, TLSHandshakeDone bool
}

// result captures all of the state after downloading URL.
type result struct {
	URL                            string
	connectionError                error
	fetchError                     error
	localAddr                      string
	resp                           *http.Response
	t0, t1, t2, t3, t4, t5, t6, t7 time.Time

	connectionTrace
}

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type transport struct {
	result  *result
	current *http.Request
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to keep track
// of the current request.
// func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
// 	t.current = req
// 	return http.DefaultTransport.RoundTrip(req)
// }

// // GotConn records local address to aid failure correlation.
// func (t *transport) GotConn(info httptrace.GotConnInfo) {
// 	t.result.localAddr = info.Conn.LocalAddr()
// }

func startWorkers(maxWorkers int, done <-chan struct{}, requests <-chan request, results chan<- *result) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case request := <-requests:
					results <- request.Fetch(request)
				case <-done:
					return
				}
			}
		}()
	}

	return wg
}

func (f fetcher) Fetch(r request) *result {
	result := &result{
		URL: r.URL,
	}

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 3 * time.Second,
	}

	// tr := &transport{
	// 	result: result,
	// }

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			result.t0 = time.Now()
			result.DNSStart = true
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			result.t1 = time.Now()
			result.DNSDone = true
		},

		ConnectStart: func(_, _ string) {
			if result.t1.IsZero() {
				result.t1 = time.Now()
			}
			result.ConnectStart = true
		},

		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				result.connectionError = fmt.Errorf("unable to connect to host %v: %v", addr, err)
			} else {
				result.ConnectDone = true
			}
			result.t2 = time.Now()
		},

		GotConn: func(i httptrace.GotConnInfo) {
			if i.Conn.LocalAddr().String() == "" {
				panic("xx")
			}
			result.localAddr = i.Conn.LocalAddr().String() // correlate with tshark
			result.t3 = time.Now()
			result.GotConn = true
		},

		GotFirstResponseByte: func() {
			result.t4 = time.Now()
			result.GotFirstResponseByte = true
		},

		TLSHandshakeStart: func() {
			result.t5 = time.Now()
			result.TLSHandshakeStart = true
		},

		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			result.t6 = time.Now()
			result.TLSHandshakeDone = true
		},
	}

	client := &http.Client{
		Timeout:   f.FetchTimeout,
		Transport: tr,
	}

	httpReq, _ := http.NewRequest("GET", r.URL, nil)
	result.resp, result.fetchError = client.Do(httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace)))

	if result.fetchError != nil {
		if result.resp != nil && result.resp.Body != nil {
			w := ioutil.Discard
			if _, err := io.Copy(w, result.resp.Body); err != nil && w != ioutil.Discard {
				log.Fatalf("failed to read response body: %v", err)
			}
			result.resp.Body.Close()
		}
		return result
	}

	result.t7 = time.Now()

	if result.t0.IsZero() {
		// we skipped DNS
		result.t0 = result.t1
	}

	return result
}

// NewHTTPFetcher returns a new Fetcher.
func NewHTTPFetcher(fetchTimeout time.Duration, maxConnectionsPerHost int) *fetcher {
	return &fetcher{
		FetchTimeout:          fetchTimeout,
		MaxConnectionsPerHost: maxConnectionsPerHost,
	}
}

func avgMillis(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}

	var total time.Duration

	for i := range values {
		total += values[i]
	}

	return total / time.Duration(len(values))
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	doneCh := make(chan struct{})
	requestCh := make(chan request)
	resultCh := make(chan *result)

	fetcher := NewHTTPFetcher(*timeout, *queue)
	wg := startWorkers(*workers, doneCh, requestCh, resultCh)

	outstandingRequests := 0
	var pending []request

	for i := 0; i < *queue; i++ {
		pending = append(pending, request{
			Fetcher: fetcher,
			URL:     flag.Arg(0),
		})

	}

	summaryCh := make(chan *result)

	go func() {
		var results []*result
		var errors []error
		var values []time.Duration
		var TLSHandshake []time.Duration

		var max time.Duration

		ticker := time.Tick(1 * time.Second)

		for {
			select {
			case result := <-summaryCh:
				d := result.TotalTime()
				if d > max {
					max = d
				}
				if *verbose {
					result.Print()
				}
				if result.fetchError != nil {
					result.Print()
					log.Printf("localAddr=%v (t3=%v), connectionError=%v, fetchError=%v\n", result.localAddr, result.t3, result.connectionError, result.fetchError)
					log.Printf("%+v", result.connectionTrace)
					cmd := exec.Command("/usr/bin/pkill", "tshark")
					log.Printf("Command finished with error: %v", cmd.Run())
					os.Exit(1)
				}
				results = append(results, result)
				if result.fetchError != nil {
					errors = append(errors, result.fetchError)
				} else {
					values = append(values, result.TotalTime())
					TLSHandshake = append(TLSHandshake, result.t6.Sub(result.t5))
				}
			case <-ticker:
				log.Printf("pending: %v --- #success: %6v, #failures: %6v, GET(avg): %v, max=%v, TLSHandshake(avg): %v",
					len(pending),
					len(values),
					len(errors),
					avgMillis(values).Round(time.Millisecond),
					max.Round(time.Millisecond),
					avgMillis(TLSHandshake).Round(time.Millisecond))
				errors = []error{}
				values = []time.Duration{}
				results = []*result{}
				TLSHandshake = []time.Duration{}
				max = 0
			}
		}
	}()

	for {
		var sendCh chan<- request
		var link request

		if len(pending) > 0 {
			sendCh = requestCh
			link = pending[0]
		} else if outstandingRequests == 0 {
			break
		}

		select {
		case sendCh <- link:
			outstandingRequests++
			pending = pending[1:]
		case result := <-resultCh:
			outstandingRequests--
			summaryCh <- result
			if result.fetchError != nil {
				if *verbose {
					log.Printf("%v", result.fetchError)
				}
			} else if result.resp != nil {
				result.resp.Body.Close()
			}
			if *repeat {
				pending = append(pending, request{
					Fetcher: fetcher,
					URL:     result.URL,
				})
			}
		}
	}

	close(doneCh)
	wg.Wait()
}
