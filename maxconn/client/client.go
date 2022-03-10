package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	for i := 0; i < 10; i++{
		go func(i int) {
			fmt.Println(i)
			resp, err := http.Get("http://localhost:8080/")
			if err != nil {
				log.Println(i, err)
				return
			}
			fmt.Println(i, resp)
		}(i)
	}

	select {}
}
