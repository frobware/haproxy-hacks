package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++{
		wg.Add(1)
		go func(i int) {
			var client = &http.Client{
				Timeout: time.Second * 10,
			}
			fmt.Println(i)
			resp, err := client.Get("http://localhost:8080/")
			if err != nil {
				log.Println(i, err)
				wg.Done()
				return
			}
			fmt.Println(i, resp)
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			fmt.Println(err)
			wg.Done()
			return
		}(i)
	}

	wg.Wait()
}
