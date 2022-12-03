package main

import (
	"fmt"
	"net"
)

func server(address string, mtu uint32) {
	s, err := net.ResolveUDPAddr("udp4", address)

	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	buffer := make([]byte, mtu)

	for {
		n, _, _ := conn.ReadFromUDP(buffer)
		fmt.Println(buffer)

		if n == 0 {
			fmt.Println("EOF")
			return
		}
	}
}
