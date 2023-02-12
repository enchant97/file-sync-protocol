package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/pbtypes"
)

func writeFromChunked(filename string, blocks map[uint64]map[uint64]core.Message) {
	// ensure we always have new file
	os.Remove(filename)

	file, _ := os.Create(filename)
	defer file.Close()

	for blockNum := 1; blockNum <= len(blocks); blockNum++ {
		chunks := blocks[uint64(blockNum)]
		for chunkNum := 1; chunkNum <= len(chunks); chunkNum++ {
			file.Write(chunks[uint64(chunkNum)].Payload)
		}
	}

}

func receivePSH(buffer []byte, conn *net.UDPConn) (core.Message, map[uint64]map[uint64]core.Message, *net.UDPAddr) {
	var clientAddress *net.UDPAddr
	var doneMessage core.Message
	queuedPayloadChunks := make([][]byte, 0)
	for {
		var n int
		n, clientAddress, _ = conn.ReadFromUDP(buffer)
		if core.PacketType(buffer[0]) == core.PacketTypeReq_PSH_DAT {
			// we don't want a reference
			dstBytes := make([]byte, n)
			copy(dstBytes, buffer[0:n])
			queuedPayloadChunks = append(queuedPayloadChunks, dstBytes)
		} else {
			strippedBuffer := buffer[0:n]
			log.Println("RX(PSH) RAW =", strippedBuffer)
			doneMessage = core.GetMessage(strippedBuffer, true)
			log.Println("RX(PSH) SER =", doneMessage)
			break
		}
	}

	// process chunks
	receivedChunks := make(map[uint64]map[uint64]core.Message)
	for _, rawChunk := range queuedPayloadChunks {
		chunk := core.GetMessage(rawChunk, true)
		header := chunk.Header.(*pbtypes.ReqPshDat)
		if _, exists := receivedChunks[header.BlockId]; !exists {
			receivedChunks[header.BlockId] = make(map[uint64]core.Message)
		}
		receivedChunks[header.BlockId][header.ChunkId] = chunk
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
	var currentRequestID uint64

	var lastReceivedMessage core.Message
	var lastSentMessage []byte

	// accept SYN
	receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)
	lastReceivedMessage = receivedMessage
	// set send mtu to match requested client's
	sendMTU := int(receivedMessage.Header.(*pbtypes.ReqSyn).MaxMtu)
	currentRequestID = receivedMessage.Header.(*pbtypes.ReqSyn).Id
	log.Printf("send MTU = '%d'\n", sendMTU)

	// send SYN back
	ackMessage, _, _ := core.MakeMessage(
		int(mtu),
		core.PacketTypeRes_SYN,
		&pbtypes.ResSyn{
			RequestId: currentRequestID,
			ClientId:  rand.Uint32(),
			MaxMtu:    mtu,
		},
		nil,
	)
	conn.WriteToUDP(ackMessage, receivedMessageAddr)
	lastSentMessage = ackMessage

	for {
		receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

		// resend last message if we receive the same message again
		if receivedMessage.Matches(lastReceivedMessage) {
			log.Println("resending last message")
			conn.WriteToUDP(lastSentMessage, receivedMessageAddr)
			continue
		}

		lastReceivedMessage = receivedMessage

		if receivedMessage.MessageType == core.PacketTypeReq_FIN {
			// FIXME handle ACK resend
			// accept REQ for FIN
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqFin).Id
			ackMessage, _, _ = core.MakeMessage(
				sendMTU,
				core.PacketTypeRes_ACK,
				&pbtypes.ResAck{
					RequestId: currentRequestID,
				},
				nil,
			)
			conn.WriteToUDP(ackMessage, receivedMessageAddr)
			break
		} else if receivedMessage.MessageType == core.PacketTypeReq_PSH {
			// accept REQ for PSH
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqPsh).Id
			pushFilePath := receivedMessage.Header.(*pbtypes.ReqPsh).Path
			pushFilePath = path.Join("./data", pushFilePath)

			// send ACK
			ackMessage, _, _ = core.MakeMessage(
				sendMTU,
				core.PacketTypeRes_ACK,
				&pbtypes.ResAck{
					RequestId: currentRequestID,
				},
				nil,
			)
			conn.WriteToUDP(ackMessage, receivedMessageAddr)
			lastSentMessage = ackMessage

			// Block ID -> Chunk ID -> Chunk
			receivedChunks := make(map[uint64]map[uint64]core.Message)
			eof := false

			// handle PSH until EOF
			log.Println("handling PSH until EOF")
			for !eof {
				// Receive PSH or REQ
				var newChunks map[uint64]map[uint64]core.Message
				receivedMessage, newChunks, receivedMessageAddr = receivePSH(buffer, conn)

				// resend last message if we receive the same message again
				if receivedMessage.Matches(lastReceivedMessage) {
					log.Println("resending last message")
					conn.WriteToUDP(lastSentMessage, receivedMessageAddr)
					continue
				}

				// add received chunks (if there are any)
				for blockID, block := range newChunks {
					for chunkID, chunk := range block {
						if _, exists := receivedChunks[blockID]; !exists {
							receivedChunks[blockID] = make(map[uint64]core.Message)
						}
						receivedChunks[blockID][chunkID] = chunk
					}
				}

				lastReceivedMessage = receivedMessage

				if receivedMessage.MessageType == core.PacketTypeReq_PSH_EOF {
					// write result to disk
					writeFromChunked(pushFilePath, receivedChunks)
					// register EOF
					eof = true
					conn.WriteToUDP(ackMessage, receivedMessageAddr)
					lastSentMessage = ackMessage
				} else if receivedMessage.MessageType == core.PacketTypeReq_PSH_VAL {
					// chunk range to validate
					blockID := receivedMessage.Header.(*pbtypes.ReqPshVal).BlockId
					lastChunkID := int(receivedMessage.Header.(*pbtypes.ReqPshVal).LastChunkId)
					subRequestID := receivedMessage.Header.(*pbtypes.ReqPshVal).SubRequestId

					// check for missing chunks
					missingChunkIDs := make([]uint64, 0)
					for chunkNum := 1; chunkNum <= lastChunkID; chunkNum++ {
						chunkNum := uint64(chunkNum)
						if _, exists := receivedChunks[blockID][chunkNum]; !exists {
							missingChunkIDs = append(missingChunkIDs, chunkNum)
						}
					}

					if len(missingChunkIDs) != 0 {
						// missing chunks were found
						log.Printf(
							"missing '%d' chunks in block '%d', expected '%d' chunks\n",
							len(missingChunkIDs), blockID, lastChunkID,
						)

						// HACK really inefficient way of handling when message is too large

						var resendMessage []byte
						chunksLenToRequest := len(missingChunkIDs)
						for {
							resendMessage, _, err = core.MakeMessage(
								sendMTU,
								core.PacketTypeRes_ERR_DAT,
								&pbtypes.ResErrDat{
									RequestId:    currentRequestID,
									SubRequestId: subRequestID,
									BlockId:      blockID,
									ChunkIds:     missingChunkIDs[:chunksLenToRequest],
								},
								nil,
							)
							if err == nil {
								break
							}
							// too many chunk id's to fit in packet, reduce by one
							chunksLenToRequest--
							log.Printf("message to big, resizing to %d\n", chunksLenToRequest)
						}
						conn.WriteToUDP(resendMessage, receivedMessageAddr)
						lastSentMessage = resendMessage
					} else {
						// no missing chunks were found
						ackMessage, _, _ = core.MakeMessage(
							sendMTU,
							core.PacketTypeRes_ACK,
							&pbtypes.ResAck{
								RequestId:    currentRequestID,
								SubRequestId: subRequestID,
							},
							nil,
						)
						conn.WriteToUDP(ackMessage, receivedMessageAddr)
						lastSentMessage = ackMessage
					}
				}
			}
			log.Println("EOF received")
		}
	}
}
