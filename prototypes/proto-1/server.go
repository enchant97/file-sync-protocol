package main

import (
	"math/rand"
	"net"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
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
	path.Join("./data", pushFilePath)

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
	for {
		receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

		if receivedMessage.MessageType == core.PacketTypeREQ {
			break
		}
	}

	// send ACK
	conn.WriteToUDP(ackMessage, receivedMessageAddr)
}
