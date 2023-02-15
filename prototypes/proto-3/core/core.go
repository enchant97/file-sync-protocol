package core

import (
	"log"
	"net"
	"time"
)

func ReceiveMessage(buffer []byte, conn *net.UDPConn, fromClient bool) (Message, *net.UDPAddr) {
	// TODO handle n=0 (connection closed)
	n, addr, _ := conn.ReadFromUDP(buffer)
	strippedBuffer := buffer[0:n]
	message := GetMessage(strippedBuffer, fromClient)
	log.Println("RX =", message.MessageType, message.Header)
	return message, addr
}

type ReceivedMessage struct {
	Message Message
	Addr    *net.UDPAddr
}

// Receives a message from a UDP connection,
// or times out after the given number of milliseconds
func ReceiveReplyOrTimeout(buffer []byte, conn *net.UDPConn, fromClient bool, timeoutMS uint) (Message, *net.UDPAddr, error) {
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeoutMS)))
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		if err, ok := err.(net.Error); !ok || !err.Timeout() {
			// not a timeout, panic
			panic(err)
		}
		return Message{}, nil, err
	}
	strippedBuffer := buffer[0:n]
	message := GetMessage(strippedBuffer, fromClient)
	log.Println("RX =", message.MessageType, message.Header)
	return message, addr, nil
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
	timeoutMS uint,
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
		receivedMessage, _, err = ReceiveReplyOrTimeout(receiveBuffer, &conn, receiveIsClient, timeoutMS)
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
