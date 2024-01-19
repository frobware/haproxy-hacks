package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an address as an argument.")
		os.Exit(1)
	}

	address := os.Args[1]

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to dial: %v\n", err)
		return
	}

	err = conn.Close()
	if err != nil {
		fmt.Printf("Failed to close the connection: %v\n", err)
	}

	fmt.Printf("Connection to %v opened and closed successfully\n", address)
}
