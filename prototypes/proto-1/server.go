package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

func writeFromChunked(filename string, chunks map[uint64]core.Message) {
	file, _ := os.Create(filename)
	defer file.Close()

	for chunkNum := 1; chunkNum <= len(chunks); chunkNum++ {
		file.Write(chunks[uint64(chunkNum)].Payload)
	}
}

func receivePSH(buffer []byte, conn *net.UDPConn) (core.Message, map[uint64]core.Message, *net.UDPAddr) {
	var clientAddress *net.UDPAddr
	var doneMessage core.Message
	queuedPayloadChunks := make([][]byte, 0)
	for {
		var n int
		n, clientAddress, _ = conn.ReadFromUDP(buffer)
		if core.PacketType(buffer[0]) == core.PacketTypePSH {
			// we don't want a reference
			dstBytes := make([]byte, n)
			copy(dstBytes, buffer[0:n])
			queuedPayloadChunks = append(queuedPayloadChunks, dstBytes)
		} else {
			strippedBuffer := buffer[0:n]
			log.Println(strippedBuffer)
			doneMessage = core.GetMessage(strippedBuffer, true)
			log.Println(doneMessage)
			break
		}
	}

	// process chunks
	receivedChunks := make(map[uint64]core.Message)
	for _, rawChunk := range queuedPayloadChunks {
		chunk := core.GetMessage(rawChunk, true)
		chunkID := chunk.Header.(*pbtypes.PshClient).ChunkId
		receivedChunks[chunkID] = chunk
	}
	return doneMessage, receivedChunks, clientAddress
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
	var receivedChunks map[uint64]core.Message
	receivedMessage, receivedChunks, receivedMessageAddr = receivePSH(buffer, conn)

	lastChunkID := int(receivedMessage.Meta.(*pbtypes.ReqPshVerifyClient).LastChunkId)

	for {
		missingChunkIDs := make([]uint64, 0)

		for chunkNum := 1; chunkNum <= lastChunkID; chunkNum++ {
			chunkNum := uint64(chunkNum)
			if _, exists := receivedChunks[chunkNum]; !exists {
				missingChunkIDs = append(missingChunkIDs, chunkNum)
			}
		}

		if len(missingChunkIDs) == 0 {
			break
		}

		log.Printf("missing '%d' chunks, expected '%d' chunks\n", len(missingChunkIDs), lastChunkID)

		resendMessage, _ := core.MakeMessage(
			int(mtu),
			core.PacketTypeREQ,
			&pbtypes.ReqServer{
				ReqId: 2,
				Type:  pbtypes.ReqTypes_REQ_RESEND_CHUNK,
			},
			&pbtypes.ReqResendChunk{
				ChunkIds: missingChunkIDs,
			},
			nil,
		)
		conn.WriteToUDP(resendMessage, receivedMessageAddr)

		_, newChunks, _ := receivePSH(buffer, conn)
		for chunkID, chunk := range newChunks {
			receivedChunks[chunkID] = chunk
		}
	}

	// send ACK
	conn.WriteToUDP(ackMessage, receivedMessageAddr)

	// write result to disk
	writeFromChunked(pushFilePath, receivedChunks)

	// receive FIN
	receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

	// send ACK
	ackMessage, _ = core.MakeMessage(
		int(mtu),
		core.PacketTypeACK,
		&pbtypes.AckServer{
			ReqId: 0,
			Type:  pbtypes.AckTypes_ACK_FIN,
		},
		nil,
		nil,
	)
	conn.WriteToUDP(ackMessage, receivedMessageAddr)
}
