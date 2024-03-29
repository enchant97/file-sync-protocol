package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"path"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-2b/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-2b/pbtypes"
)

func writeFromChunked(filename string, chunks map[uint64]core.Message) {
	// ensure we always have new file
	os.Remove(filename)

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
	currentRequestID := receivedMessage.Header.(*pbtypes.SynClient).Id
	// set send mtu to match requested client's
	sendMTU := int(receivedMessage.Header.(*pbtypes.SynClient).Mtu)
	log.Printf("send MTU = '%d'\n", sendMTU)

	// send ACK
	ackMessage, _, _ := core.MakeMessage(
		int(mtu),
		core.PacketTypeACK,
		&pbtypes.AckServer{
			ReqId: currentRequestID,
			Type:  pbtypes.AckTypes_ACK_SYN,
		},
		&pbtypes.AckSynServer{
			ClientId: rand.Uint32(),
			Mtu:      mtu,
		},
		nil,
	)
	conn.WriteToUDP(ackMessage, receivedMessageAddr)

	for {
		receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

		// handle REQ messages, ignoring unknown
		if receivedMessage.MessageType == core.PacketTypeFIN {
			// accept REQ for FIN
			currentRequestID = receivedMessage.Header.(*pbtypes.FinClient).Id
			// send ACK
			ackMessage, _, _ = core.MakeMessage(
				sendMTU,
				core.PacketTypeACK,
				&pbtypes.AckServer{
					ReqId: currentRequestID,
					Type:  pbtypes.AckTypes_ACK_FIN,
				},
				nil,
				nil,
			)
			conn.WriteToUDP(ackMessage, receivedMessageAddr)
			break
		} else if receivedMessage.MessageType == core.PacketTypeREQ {
			// accept REQ for PSH
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqClient).Id
			pushFilePath := receivedMessage.Meta.(*pbtypes.ReqPshClient).Path
			pushFilePath = path.Join("./data", pushFilePath)

			// send ACK
			ackMessage, _, _ = core.MakeMessage(
				sendMTU,
				core.PacketTypeACK,
				&pbtypes.AckServer{
					ReqId: currentRequestID,
					Type:  pbtypes.AckTypes_ACK_REQ,
				},
				nil,
				nil,
			)
			conn.WriteToUDP(ackMessage, receivedMessageAddr)

			receivedChunks := make(map[uint64]core.Message)
			eof := false

			// handle PSH until EOF
			for !eof {
				// Receive PSH or REQ
				var newChunks map[uint64]core.Message
				receivedMessage, newChunks, receivedMessageAddr = receivePSH(buffer, conn)

				// add received chunks (if there are any)
				for chunkID, chunk := range newChunks {
					receivedChunks[chunkID] = chunk
				}

				messageType := receivedMessage.Header.(*pbtypes.ReqClient).Type
				if messageType == pbtypes.ReqTypes_REQ_PUSH_EOF {
					// write result to disk
					writeFromChunked(pushFilePath, receivedChunks)
					// register EOF
					eof = true
					conn.WriteToUDP(ackMessage, receivedMessageAddr)
				} else if messageType == pbtypes.ReqTypes_REQ_PUSH_VERIFY {
					// chunk range to validate
					lastChunkID := int(receivedMessage.Meta.(*pbtypes.ReqPshVerifyClient).LastChunkId)

					// check for missing chunks
					missingChunkIDs := make([]uint64, 0)
					for chunkNum := 1; chunkNum <= lastChunkID; chunkNum++ {
						chunkNum := uint64(chunkNum)
						if _, exists := receivedChunks[chunkNum]; !exists {
							missingChunkIDs = append(missingChunkIDs, chunkNum)
						}
					}

					if len(missingChunkIDs) != 0 {
						// missing chunks were found
						log.Printf("missing '%d' chunks, expected '%d' chunks\n", len(missingChunkIDs), lastChunkID)

						// HACK really inefficient way of handling when message is too large

						var resendMessage []byte
						chunksLenToRequest := len(missingChunkIDs)
						for {
							resendMessage, _, err = core.MakeMessage(
								sendMTU,
								core.PacketTypeREQ,
								&pbtypes.ReqServer{
									ReqId: currentRequestID,
									Type:  pbtypes.ReqTypes_REQ_RESEND_CHUNK,
								},
								&pbtypes.ReqResendChunk{
									ChunkIds: missingChunkIDs[:chunksLenToRequest],
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
					} else {
						// no missing chunks were found
						conn.WriteToUDP(ackMessage, receivedMessageAddr)
					}
				}
			}
		}
	}
}
