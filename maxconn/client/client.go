package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const N int = 300

func main() {
	var wg sync.WaitGroup
	var successful int

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer func() {
				wg.Done()
			}()
			var client = &http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := client.Get(fmt.Sprintf("http://localhost:8080/id=%d", i))
			if err != nil {
				log.Println("Connection failed", i)
				return
			}
			defer resp.Body.Close()
			if _, err := ioutil.ReadAll(resp.Body); err != nil {
				fmt.Println("read error", i, err)
			} else {
				successful +=1
				fmt.Println(i, resp.StatusCode)
			}
			return
		}(i)
	}

	wg.Wait()
	fmt.Println("successful: ", successful)
}