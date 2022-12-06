package core

import (
	"fmt"
	"net"
)

func ReceiveMessage(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr) {
	n, addr, _ := conn.ReadFromUDP(buffer)
	strippedBuffer := buffer[0:n]
	fmt.Println(strippedBuffer)
	message := GetMessage(strippedBuffer, fromClient)
	fmt.Println(message)
	return message, addr
}
