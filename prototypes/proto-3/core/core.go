package core

import (
	"fmt"
	"log"
	"net"
	"time"
)

const ReceiveTimeoutMS = 1000

func ReceiveMessage(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr) {
	// TODO handle n=0 (connection closed)
	n, addr, _ := conn.ReadFromUDP(buffer)
	strippedBuffer := buffer[0:n]
	log.Println("RX RAW =", strippedBuffer)
	message := GetMessage(strippedBuffer, fromClient)
	log.Println("RX DES =", message)
	return message, addr
}

type ReceivedMessage struct {
	Message Message
	Addr    *net.UDPAddr
}

// Receives a message from a UDP connection,
// or times out after the given number of milliseconds
func ReceiveReplyOrTimeout(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr, error) {
	receivedMessage := make(chan ReceivedMessage, 1)
	go func() {
		message, addr := ReceiveMessage(buffer, conn, fromClient)
		receivedMessage <- ReceivedMessage{message, addr}
	}()
	select {
	case <-time.After(time.Millisecond * time.Duration(ReceiveTimeoutMS)):
		// Timeout occurred
		return Message{}, nil, fmt.Errorf("timed out waiting for reply")
	case message := <-receivedMessage:
		// Message received ok
		return message.Message, message.Addr, nil
	}
}

// Sends a request and waits for a reply,
// resending the request if it times out
func SendAndReceiveRequest(
	sendMTU int,
	message []byte,
	receivePacketType PacketType,
	receiveIsClient bool,
	replyRequestID uint64,
	receiveBuffer []byte,
	conn net.UDPConn,
	sendAddr *net.UDPAddr,
) Message {
	// receive reply
	var receivedMessage Message
	var err error
	if sendAddr == nil {
		// send request as client
		conn.Write(message)
		log.Println("TX RAW =", message)
	} else {
		// send request as server
		log.Println("TX RAW =", message)
		conn.WriteToUDP(message, sendAddr)
	}
	for {
		receivedMessage, _, err = ReceiveReplyOrTimeout(receiveBuffer, &conn, receiveIsClient)
		if err != nil {
			// timeout, resend
			log.Println("timeout, resending request")
			if sendAddr == nil {
				// send request as client
				conn.Write(message)
			} else {
				// send request as server
				conn.WriteToUDP(message, sendAddr)
			}
		} else if receivePacketType != 0 && receivedMessage.MessageType != receivePacketType {
			// not the reply we want, ignore
			log.Println("received message with wrong packet type, ignoring")
		} else if receivedMessage.Header.ProtoReflect().Descriptor().Fields().ByName("request_id") == nil {
			// not the reply we want, ignore
			log.Println("received message with no request ID, ignoring")
		} else if receivedMessage.Header.ProtoReflect().Get(receivedMessage.Header.ProtoReflect().Descriptor().Fields().ByName("request_id")).Uint() != replyRequestID {
			// not the reply we want, ignore
			log.Println("received message with wrong request ID, ignoring")
		} else {
			// got the reply we want
			break
		}
	}
	return receivedMessage
}
