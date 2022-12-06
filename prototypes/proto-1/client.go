package main

import (
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

func client(address string, mtu uint32, filePath string) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, mtu)

	// var receivedMessage core.Message

	// send SYN
	messageToSend, _ := core.MakeMessage(
		int(mtu),
		core.PacketTypeSYN,
		&pbtypes.SynClient{
			Id:  1,
			Mtu: mtu,
		},
		nil,
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	core.ReceiveMessage(buffer, conn, false)

	// send Req for PSH
	fileInfo, _ := os.Stat(filePath)
	messageToSend, _ = core.MakeMessage(
		int(mtu),
		core.PacketTypeREQ,
		&pbtypes.ReqClient{
			Id:   2,
			Type: pbtypes.ReqTypes_PUSH_OBJ,
		},
		&pbtypes.ReqPshClient{
			Path: "from-client" + path.Ext(filePath),
			Size: uint64(fileInfo.Size()),
		},
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	core.ReceiveMessage(buffer, conn, false)
}
