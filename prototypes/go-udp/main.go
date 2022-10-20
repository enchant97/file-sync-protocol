package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	if args[1] == "server" {
		server(args[2], args[3])
	}
	if args[1] == "client" {
		client(args[2], args[3])
	}
	fmt.Println("done")
}
