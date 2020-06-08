package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Fetcher fetches HTML documents.
type Fetcher interface {
	// Fetch returns a reader for the body of the downloaded URL,
	// or error if it could not be downloaded. The caller is
	// responsible for body.Close().
	Fetch(url string) (body io.ReadCloser, err error)
}

// Ensure fetcher is a Fetcher.
var _ Fetcher = (*fetcher)(nil)

type fetcher struct {
	FetchTimeout time.Duration
	Fetcher
}

type request struct {
	Fetcher
	URL string
}

// result captures all of the state after downloading request.URL.
type result struct {
	request
	getDuration time.Duration
	error       error
}

func fetch(req request) *result {
	startTime := time.Now()
	result := &result{request: req}
	body, err := req.Fetcher.Fetch(req.URL)

	if err != nil {
		result.error = err
		return result
	}

	defer body.Close()
	result.getDuration = time.Now().Sub(startTime)
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
				case req := <-requests:
					results <- fetch(req)
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
func (f fetcher) Fetch(URL string) (io.ReadCloser, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Timeout: f.FetchTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
		},
	}

	if f.FetchTimeout > 0 {
		client.Timeout = f.FetchTimeout
	}

	resp, err := client.Get(URL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("fetch failed: HTTP status %d", resp.StatusCode)
		return nil, errors.New(msg)
	}

	return resp.Body, nil
}

// NewHTTPFetcher returns a new Fetcher.
func NewHTTPFetcher(fetchTimeout time.Duration) *fetcher {
	return &fetcher{
		FetchTimeout: fetchTimeout,
	}
}

var (
	verbose        = flag.Bool("v", false, "Verbose")
	workers        = flag.Int("workers", 50, "number of GET workers")
	timeout        = flag.Duration("timeout", 100*time.Millisecond, "GET timeout")
	concurrentGets = flag.Int("c", 100, "concurrent GET requests")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	doneCh := make(chan struct{})
	requestCh := make(chan request)
	resultCh := make(chan *result)

	fetcher := NewHTTPFetcher(*timeout)
	wg := startWorkers(*workers, doneCh, requestCh, resultCh)

	var pending []request

	for i := 0; i < *concurrentGets; i++ {
		pending = append(pending, request{
			Fetcher: fetcher,
			URL:     flag.Arg(0),
		})

	}

	successCh := make(chan time.Duration)

	go func() {
		var n int
		values := make([]int64, *workers, *workers)
		ticker := time.Tick(3 * time.Second)

		for {
			select {
			case duration := <-successCh:
				values[n%len(values)] = duration.Milliseconds()
				n = n + 1
			case <-ticker:
				var total int64 = 0
				for i := range values {
					total += values[i]
				}
				var avg int64 = total / int64(len(values))
				log.Println("Avg GET", time.Duration(avg)*time.Millisecond)
			}
		}
	}()

	outstandingFetches := 0

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
			if result.error != nil {
				log.Println(fetcher.FetchTimeout, result.error.Error())
			} else {
				successCh <- result.getDuration
			}
			pending = append(pending, request{
				Fetcher: result.Fetcher,
				URL:     result.URL,
			})
		}
	}

	close(doneCh)
	wg.Wait()
}
