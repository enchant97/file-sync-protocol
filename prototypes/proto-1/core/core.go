package core

import (
	"fmt"
	"net"
)

func ReceiveMessage(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr) {
	n, addr, _ := conn.ReadFromUDP(buffer)
	fmt.Println(buffer)
	message := GetMessage(buffer[0:n], fromClient)
	fmt.Println(message)
	return message, addr
}
