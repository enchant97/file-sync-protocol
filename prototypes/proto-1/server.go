package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

func writeFromChunked(filename string, messageChunks []core.Message) {
	file, _ := os.Create(filename)
	defer file.Close()

	for _, chunk := range messageChunks {
		file.Write(chunk.Payload)
	}
}

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
	var receivedMessage core.Message
	var receivedMessageAddr *net.UDPAddr

	// accept SYN
	receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

	// send ACK
	ackMessage, _ := core.MakeMessage(
		int(mtu),
		core.PacketTypeACK,
		&pbtypes.AckServer{
			ReqId: 1,
			Type:  pbtypes.AckTypes_ACK_SYN,
		},
		&pbtypes.AckSynServer{
			ClientId: rand.Uint32(),
			Mtu:      mtu,
		},
		nil,
	)
	conn.WriteToUDP(ackMessage, receivedMessageAddr)

	// accept REQ for PSH
	receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)
	pushFilePath := receivedMessage.Meta.(*pbtypes.ReqPshClient).Path
	pushFilePath = path.Join("./data", pushFilePath)

	// send ACK
	ackMessage, _ = core.MakeMessage(
		int(mtu),
		core.PacketTypeACK,
		&pbtypes.AckServer{
			ReqId: 2,
			Type:  pbtypes.AckTypes_ACK_REQ,
		},
		nil,
		nil,
	)
	conn.WriteToUDP(ackMessage, receivedMessageAddr)

	// accept pushed chunks
	queuedPayloadChunks := make([][]byte, 0)
	for {
		var n int
		n, receivedMessageAddr, _ = conn.ReadFromUDP(buffer)
		if core.PacketType(buffer[0]) == core.PacketTypePSH {
			// we don't want a reference
			dstBytes := make([]byte, n)
			copy(dstBytes, buffer[0:n])
			queuedPayloadChunks = append(queuedPayloadChunks, dstBytes)
		} else {
			strippedBuffer := buffer[0:n]
			fmt.Println(strippedBuffer)
			receivedMessage = core.GetMessage(strippedBuffer, true)
			fmt.Println(receivedMessage)
			break
		}
	}

	receivedMessages := make([]core.Message, len(queuedPayloadChunks))
	for i := 0; i < len(queuedPayloadChunks); i++ {
		receivedMessages[i] = core.GetMessage(queuedPayloadChunks[i], true)
	}

	// write result to disk in "background"
	go writeFromChunked(pushFilePath, receivedMessages)

	// send ACK
	conn.WriteToUDP(ackMessage, receivedMessageAddr)

}
