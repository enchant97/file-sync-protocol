package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	mtu := uint32(512)

	if args[1] == "server" {
		server(args[2], mtu)
	}
	if args[1] == "client" {
		client(args[2], mtu, args[3])
	}
	fmt.Println("done")
}
