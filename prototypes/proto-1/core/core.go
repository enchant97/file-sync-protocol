package core

import (
	"log"
	"net"
)

func ReceiveMessage(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr) {
	n, addr, _ := conn.ReadFromUDP(buffer)
	strippedBuffer := buffer[0:n]
	log.Println(strippedBuffer)
	message := GetMessage(strippedBuffer, fromClient)
	log.Println(message)
	return message, addr
}
