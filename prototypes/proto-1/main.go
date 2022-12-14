package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args

	mtu := uint32(512)

	if value, isSet := os.LookupEnv("NET_MTU"); isSet {
		value, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			log.Fatalln("NET_MTU invalid")
		}
		mtu = uint32(value)
	}

	log.Printf("receive MTU = '%d'\n", mtu)

	if args[1] == "server" {
		log.Println("starting server...")
		server(args[2], mtu)
	}
	if args[1] == "client" {
		log.Println("starting client...")
		client(args[2], mtu, args[3])
	}
	log.Println("done")
}
