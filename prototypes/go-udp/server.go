package main

import (
	"fmt"
	"net"
	"os"
)

func server(address string, filepathOut string) {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", s)
	if err != nil {
		panic(err)
	}

	f, _ := os.Create(filepathOut)
	defer f.Close()

	defer conn.Close()
	buffer := make([]byte, 8192)

	okData := []byte("ok")

	for {
		n, addr, _ := conn.ReadFromUDP(buffer)

		if n == 0 {
			fmt.Println("EOF")
			return
		}

		f.Write(buffer[0:n])

		conn.WriteToUDP(okData, addr)
	}
}
