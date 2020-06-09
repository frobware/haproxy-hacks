package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Fetcher fetches HTML documents.
type Fetcher interface {
	// Fetch returns a reader for the body of the downloaded URL,
	// or error if it could not be downloaded. The caller is
	// responsible for body.Close().
	Fetch(url string) (*http.Response, error)
}

// Ensure fetcher is a Fetcher.
var _ Fetcher = (*fetcher)(nil)

var (
	verbose = flag.Bool("v", false, "Verbose")
	workers = flag.Int("workers", 50, "number of GET workers")
	timeout = flag.Duration("timeout", 100*time.Millisecond, "GET timeout")
	queue   = flag.Int("queue", 100, "queue <N> GET requests")
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

// result captures all of the state after downloading request.URL.
type result struct {
	request
	startTime  time.Time
	endTime    time.Time
	fetchError error
	resp       *http.Response
}

func fetch(req request) *result {
	result := &result{
		request:   req,
		startTime: time.Now(),
	}
	result.resp, result.fetchError = req.Fetcher.Fetch(req.URL)
	result.endTime = time.Now()
	return result
}

func startWorkers(maxWorkers int, done <-chan struct{}, requests <-chan request, results chan<- *result) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case request := <-requests:
					results <- fetch(request)
				case <-done:
					return
				}
			}
		}()
	}

	return wg
}

// Fetch URL returning the reader to the body of the document, or an
// error if URL could not be fetched. The caller must call Close() on
// the reader to avoid resource leaks.
func (f fetcher) Fetch(URL string) (*http.Response, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Timeout: f.FetchTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
			// Proxy:           http.ProxyFromEnvironment,
			// DialContext: (&net.Dialer{
			// 	Timeout: f.FetchTimeout,
			// }).DialContext,
			// MaxIdleConnsPerHost: f.MaxConnectionsPerHost,
		},
	}

	if f.FetchTimeout > 0 {
		client.Timeout = f.FetchTimeout
	}

	return client.Get(URL)
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

	summaryCh := make(chan *result)

	go func() {
		var results []*result
		var errors []error
		var values []time.Duration

		var max time.Duration

		ticker := time.Tick(1 * time.Second)

		for {
			select {
			case result := <-summaryCh:
				d := result.endTime.Sub(result.startTime)
				if d > max {
					max = d
				}
				if result.fetchError != nil {
					log.Printf("start=%v, end=%v, duration=%v\n", result.startTime, result.endTime, result.endTime.Sub(result.startTime))
					log.Printf("resp=%+v, error=%v\n", result.resp, result.fetchError)
					os.Exit(1)
				}
				results = append(results, result)
				if result.fetchError != nil {
					errors = append(errors, result.fetchError)
				} else {
					values = append(values, result.endTime.Sub(result.startTime))
				}
			case <-ticker:
				log.Printf("#success: %6v, #failures: %6v, GET(avg): %v, max=%v",
					len(values),
					len(errors),
					avgMillis(values).Round(time.Millisecond),
					max.Round(time.Millisecond))
				if len(errors) > 0 {
					for k, v := range results {
						log.Printf("%d: start=%v, end=%v, duration=%v\n", k, v.startTime, v.endTime, v.endTime.Sub(v.startTime))
						break
					}
					os.Exit(1)
				}
				errors = []error{}
				values = []time.Duration{}
				results = []*result{}
			}
		}
	}()

	outstandingFetches := 0
	var pending []request

	for i := 0; i < *queue; i++ {
		pending = append(pending, request{
			Fetcher: fetcher,
			URL:     flag.Arg(0),
		})

	}

	for {
		var sendCh chan<- request
		var link request

		if len(pending) > 0 {
			sendCh = requestCh
			link = pending[0]
		} else if outstandingFetches == 0 {
			break
		}

		select {
		case sendCh <- link:
			outstandingFetches++
			pending = pending[1:]
		case result := <-resultCh:
			outstandingFetches--
			summaryCh <- result
			if result.fetchError != nil {
				if *verbose {
					log.Printf("%v", result.fetchError)
				}
			} else if result.resp != nil {
				result.resp.Body.Close()
			}
			if len(pending) < *queue {
				pending = append(pending, request{
					Fetcher: result.Fetcher,
					URL:     result.URL,
				})
			}
		}
	}

	close(doneCh)
	wg.Wait()
}
