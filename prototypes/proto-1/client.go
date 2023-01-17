package main

import (
	"log"
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

// The 'safe' mtu payload size
// https://stackoverflow.com/a/1099359
const SYN_MTU_SIZE = 512 - core.Ipv4ReservedBytes

func client(address string, mtu uint32, filePath string) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, mtu)

	var receivedMessage core.Message

	// send SYN
	messageToSend, _, _ := core.MakeMessage(
		SYN_MTU_SIZE,
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
	receivedMessage, _ = core.ReceiveMessage(buffer, conn, false)
	// set send mtu to match requested server's
	sendMTU := int(receivedMessage.Meta.(*pbtypes.AckSynServer).Mtu)
	log.Printf("send MTU = '%d'\n", sendMTU)

	// send Req for PSH
	fileInfo, _ := os.Stat(filePath)
	messageToSend, _, _ = core.MakeMessage(
		sendMTU,
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
	var seekOffset int = 0
	var chunkIDToOffset = make(map[uint64]int)
	for {
		// send next chunk
		payloadMessageToSend, payloadLength, _ := core.MakeMessage(
			sendMTU,
			core.PacketTypePSH,
			&pbtypes.PshClient{
				ReqId:   2,
				ChunkId: uint64(lastChunkID) + 1,
			},
			nil,
			fileReader,
		)
		if payloadLength == 0 {
			// EOF
			break
		}
		lastChunkID += 1
		chunkIDToOffset[lastChunkID] = seekOffset
		seekOffset += payloadLength
		conn.Write(payloadMessageToSend)
	}

	for {
		// send REQ verify
		messageToSend, _, _ = core.MakeMessage(
			sendMTU,
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
		receivedMessage, _ = core.ReceiveMessage(buffer, conn, false)
		if receivedMessage.MessageType == core.PacketTypeACK {
			break
		}

		// receive REQ for resend
		missingChunks := receivedMessage.Meta.(*pbtypes.ReqResendChunk).ChunkIds

		// send the requested missing chunks
		for _, chunkID := range missingChunks {
			fileReader.Seek(int64(chunkIDToOffset[chunkID]), 0)
			payloadMessageToSend, _, _ := core.MakeMessage(
				sendMTU,
				core.PacketTypePSH,
				&pbtypes.PshClient{
					ReqId:   2,
					ChunkId: chunkID,
				},
				nil,
				fileReader,
			)
			conn.Write(payloadMessageToSend)
		}
	}

	// send FIN
	messageToSend, _, _ = core.MakeMessage(
		sendMTU,
		core.PacketTypeFIN,
		nil,
		nil,
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	core.ReceiveMessage(buffer, conn, false)
}
