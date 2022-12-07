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
			Type: pbtypes.ReqTypes_REQ_PUSH_OBJ,
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

	// push all chunks
	fileReader, fileReadErr := os.Open(filePath)
	if fileReadErr != nil {
		panic(fileReadErr)
	}
	defer fileReader.Close()

	var lastChunkID uint64 = 0
	for {
		// send next chunk
		payloadMessageToSend, payloadLength := core.MakeMessage(
			int(mtu),
			core.PacketTypePSH,
			&pbtypes.PshClient{
				ReqId:   2,
				ChunkId: uint64(lastChunkID),
			},
			nil,
			fileReader,
		)
		if payloadLength == 0 {
			// EOF
			break
		}
		lastChunkID += 1
		conn.Write(payloadMessageToSend)
	}

	// send REQ verify
	messageToSend, _ = core.MakeMessage(
		int(mtu),
		core.PacketTypeREQ,
		&pbtypes.ReqClient{
			Id:   2,
			Type: pbtypes.ReqTypes_REQ_PSH_VERIFY,
		},
		&pbtypes.ReqPshVerifyClient{
			LastChunkId: lastChunkID,
		},
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	core.ReceiveMessage(buffer, conn, false)
}
