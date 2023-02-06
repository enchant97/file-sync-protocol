package main

import (
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/pbtypes"
)

// The 'safe' mtu payload size
// https://stackoverflow.com/a/1099359
const SYN_MTU_SIZE = 512 - core.Ipv4ReservedBytes

func client(address string, mtu uint32, chunks_per_block uint, filePaths []string) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, mtu)

	var receivedMessage core.Message
	var currentRequestID uint64 = 1

	// send SYN
	messageToSend, _, _ := core.MakeMessage(
		SYN_MTU_SIZE,
		core.PacketTypeReq_SYN,
		&pbtypes.ReqSyn{
			Id:     currentRequestID,
			MaxMtu: mtu,
		},
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	receivedMessage, _ = core.ReceiveMessage(buffer, conn, false)
	// set send mtu to match requested server's
	sendMTU := int(receivedMessage.Header.(*pbtypes.ResSyn).MaxMtu)
	log.Printf("send MTU = '%d'\n", sendMTU)

	for _, filePath := range filePaths {
		// send Req for PSH
		currentRequestID++
		fileInfo, _ := os.Stat(filePath)
		messageToSend, _, _ = core.MakeMessage(
			sendMTU,
			core.PacketTypeReq_PSH,
			&pbtypes.ReqPsh{
				Id:   currentRequestID,
				Path: path.Base(filePath),
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

		var currentBlockID uint64 = 0
		var lastChunkID uint64 = 0
		var seekOffset int = 0
		// Chunk ID -> Seek Offset
		chunkIDToOffset := make(map[uint64]int)
		// Remove all recorded offsets, used when new block is loaded
		clearChunkIDOffsets := func() {
			for k := range chunkIDToOffset {
				delete(chunkIDToOffset, k)
			}
		}

		createChunkMessage := func(chunkID uint64) ([]byte, int, error) {
			return core.MakeMessage(
				sendMTU,
				core.PacketTypeReq_PSH_DAT,
				&pbtypes.ReqPshDat{
					RequestId: currentRequestID,
					BlockId:   currentBlockID,
					ChunkId:   chunkID,
				},
				fileReader,
			)
		}

		eof := false
		// missing chunk id's for current block
		missingChunks := make([]uint64, 0)

		for {
			if len(missingChunks) == 0 && eof {
				break
			}
			if len(missingChunks) != 0 {
				// PSH missing
				for _, chunkID := range missingChunks {
					fileReader.Seek(int64(chunkIDToOffset[chunkID]), 0)
					payloadMessageToSend, _, _ := createChunkMessage(chunkID)
					conn.Write(payloadMessageToSend)
				}
				log.Printf("PSH requested '%d' missing chunks", len(missingChunks))
				missingChunks = nil
			} else {
				// PSH next block
				// NOTE this ensures file position is reset to last position if a resend was issued
				fileReader.Seek(int64(seekOffset), 0)
				currentBlockID++
				lastChunkID = 0
				clearChunkIDOffsets()

				chunksSent := uint(0)
				for chunksSent < chunks_per_block && !eof {
					payloadMessageToSend, payloadLength, _ := createChunkMessage(uint64(lastChunkID) + 1)
					if payloadLength == 0 {
						eof = true
					} else {
						lastChunkID += 1
						chunkIDToOffset[lastChunkID] = seekOffset
						seekOffset += payloadLength
						conn.Write(payloadMessageToSend)
						chunksSent++
					}
				}
			}

			// HACK server cannot keep up
			time.Sleep(time.Millisecond * 2)

			// send REQ verify
			messageToSend, _, _ = core.MakeMessage(
				sendMTU,
				core.PacketTypeReq_PSH_VAL,
				&pbtypes.ReqPshVal{
					RequestId:   currentRequestID,
					BlockId:     currentBlockID,
					LastChunkId: lastChunkID,
				},
				nil,
			)
			conn.Write(messageToSend)

			// receive ACK or REQ for resend
			receivedMessage, _ = core.ReceiveMessage(buffer, conn, false)
			if receivedMessage.MessageType != core.PacketTypeRes_ACK {
				missingChunks = append(missingChunks, receivedMessage.Header.(*pbtypes.ResErrDat).ChunkIds...)
				log.Printf("REQ for '%d' missing chunks", len(missingChunks))
			}
		}

		// send EOF
		messageToSend, _, _ = core.MakeMessage(
			sendMTU,
			core.PacketTypeReq_PSH_EOF,
			&pbtypes.ReqPshEof{
				RequestId: currentRequestID,
			},
			nil,
		)
		conn.Write(messageToSend)

		// receive ACK
		core.ReceiveMessage(buffer, conn, false)
	}

	// send FIN
	currentRequestID++
	messageToSend, _, _ = core.MakeMessage(
		sendMTU,
		core.PacketTypeReq_FIN,
		&pbtypes.ReqFin{
			Id: currentRequestID,
		},
		nil,
	)
	conn.Write(messageToSend)

	// receive ACK
	core.ReceiveMessage(buffer, conn, false)
}
