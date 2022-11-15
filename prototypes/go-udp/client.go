package main

import (
	"fmt"
	"net"
	"os"
)

func client(address string, filepath string) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}

	f, _ := os.Open(filepath)

	defer f.Close()
	defer conn.Close()

	writeBuffer := make([]byte, 8192)

	for {
		writeN, _ := f.Read(writeBuffer)
		if writeN == 0 {
			fmt.Println("EOF")
			conn.Write([]byte(""))
			return
		}
		if _, err := conn.Write(writeBuffer[0:writeN]); err != nil {
			panic(err)
		}
	}
}
