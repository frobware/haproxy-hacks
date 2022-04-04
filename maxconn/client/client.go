package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const N int = 300

var (
	iterations = flag.Int("iterations", 100, "iterations")
)

func pokeURL(url string) {
	var wg sync.WaitGroup
	var successful int

	for i := 0; i < *iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer func() {
				wg.Done()
			}()
			var client = &http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := client.Get(fmt.Sprintf("http://%s?id=%d", url, i))
			if err != nil {
				log.Println(url, "Connection failed", i, err)
				return
			}
			defer resp.Body.Close()
			if _, err := ioutil.ReadAll(resp.Body); err != nil {
				log.Println(url, "read error", i, err)
			} else {
				successful += 1
				log.Println(i, url, resp.StatusCode)
			}
			return
		}(i)
	}

	wg.Wait()
	log.Println("successful: ", successful)
}

func main() {
	pokeURL("localhost:8080/1")
	pokeURL("localhost:8081/2")
}
