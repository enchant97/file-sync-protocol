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

type CurrentFile struct {
	// FilePath
	FilePath string
	// Block ID -> Chunk ID -> Chunk
	Blocks map[uint64]map[uint64][]byte
}

func (cf *CurrentFile) AddNewChunk(blockID uint64, chunkID uint64, chunk []byte) {
	if _, exists := cf.Blocks[blockID]; !exists {
		cf.Blocks[blockID] = make(map[uint64][]byte)
	}
	// we don't want a reference
	dstBytes := make([]byte, len(chunk))
	copy(dstBytes, chunk)
	cf.Blocks[blockID][chunkID] = dstBytes
}

func (cf *CurrentFile) FindMissingChunksForBlock(blockID uint64, lastChunkID uint64) []uint64 {
	missingChunks := make([]uint64, 0)
	for chunkID := uint64(1); chunkID <= lastChunkID; chunkID++ {
		if _, exists := cf.Blocks[blockID][chunkID]; !exists {
			missingChunks = append(missingChunks, chunkID)
		}
	}
	return missingChunks
}

func (cf *CurrentFile) WriteToFile() {
	// ensure we always have new file
	os.Remove(cf.FilePath)

	file, _ := os.Create(cf.FilePath)
	defer file.Close()

	for blockNum := 1; blockNum <= len(cf.Blocks); blockNum++ {
		chunks := cf.Blocks[uint64(blockNum)]
		for chunkNum := 1; chunkNum <= len(chunks); chunkNum++ {
			file.Write(chunks[uint64(chunkNum)])
		}
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
	sendMTU := int(mtu)
	var receivedMessage core.Message
	var receivedMessageAddr *net.UDPAddr
	var currentRequestID uint64
	var currentSubRequestID uint64 = 0

	var lastReceivedMessage core.Message
	var lastSentMessage []byte

	var currentFile CurrentFile

	sendMessage := func(message []byte) {
		conn.WriteToUDP(message, receivedMessageAddr)
		lastSentMessage = message
	}

	sendSynMessage := func() {
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
		sendMessage(ackMessage)
	}

	sendAckMessage := func(requestID uint64, subRequestID uint64) {
		ackMessage, _, _ := core.MakeMessage(
			int(mtu),
			core.PacketTypeRes_ACK,
			&pbtypes.ResAck{
				RequestId:    currentRequestID,
				SubRequestId: subRequestID,
			},
			nil,
		)
		sendMessage(ackMessage)
	}

	sendErrDatMessage := func(requestID uint64, subRequestID uint64, blockID uint64, missingChunkIDs []uint64) {
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
		sendMessage(resendMessage)
	}

	// receive messages from a single client
	fin := false
	for !fin {
		receivedMessage, receivedMessageAddr = core.ReceiveMessage(buffer, conn, true)

		// resend last message if we receive the same message again
		if receivedMessage.Matches(lastReceivedMessage) {
			log.Println("resending last message")
			sendMessage(lastSentMessage)
			continue
		}

		// check if message is from the past
		if descriptor := receivedMessage.Header.ProtoReflect().Descriptor().Fields().ByNumber(1); descriptor != nil && receivedMessage.Header.ProtoReflect().Get(descriptor).Uint() < currentRequestID {
			log.Println("received message from the past, ignoring")
			continue
		}
		// FIXME validate sub request id

		// update last received message
		lastReceivedMessage = receivedMessage

		// handle a new (unseen) message
		switch receivedMessage.MessageType {
		case core.PacketTypeReq_SYN:
			currentSubRequestID = 0
			// Handle REQ for SYN
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqSyn).Id
			// set send mtu to match requested client's
			sendMTU = int(receivedMessage.Header.(*pbtypes.ReqSyn).MaxMtu)
			log.Printf("send MTU = '%d'\n", sendMTU)
			sendSynMessage()
		case core.PacketTypeReq_FIN:
			currentSubRequestID = 0
			// Handle REQ for FIN
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqFin).Id
			// FIXME handle ACK resend
			sendAckMessage(currentRequestID, 0)
			fin = true
		case core.PacketTypeReq_PSH:
			currentSubRequestID = 0
			// Handle REQ for PSH
			currentRequestID = receivedMessage.Header.(*pbtypes.ReqPsh).Id
			filePath := receivedMessage.Header.(*pbtypes.ReqPsh).Path
			filePath = path.Join("./data", filePath)
			currentFile = CurrentFile{
				FilePath: filePath,
				Blocks:   make(map[uint64]map[uint64][]byte),
			}
			sendAckMessage(currentRequestID, 0)
		case core.PacketTypeReq_PSH_DAT:
			// Handle PSH-DAT chunks
			currentBlockID := receivedMessage.Header.(*pbtypes.ReqPshDat).BlockId
			currentChunkID := receivedMessage.Header.(*pbtypes.ReqPshDat).ChunkId
			currentChunk := receivedMessage.Payload
			currentFile.AddNewChunk(currentBlockID, currentChunkID, currentChunk)
			sendAckMessage(currentRequestID, currentChunkID)
		case core.PacketTypeReq_PSH_VAL:
			// Handle PSH-VAL
			currentSubRequestID = receivedMessage.Header.(*pbtypes.ReqPshVal).SubRequestId
			currentBlockID := receivedMessage.Header.(*pbtypes.ReqPshVal).BlockId
			lastChunkID := receivedMessage.Header.(*pbtypes.ReqPshVal).LastChunkId
			missingChunks := currentFile.FindMissingChunksForBlock(currentBlockID, lastChunkID)
			if len(missingChunks) > 0 {
				// missing chunks, send ERR-DAT for requesting missing chunks
				log.Printf(
					"missing '%d' chunks in block '%d', expected '%d' chunks\n",
					len(missingChunks), currentBlockID, lastChunkID,
				)
				sendErrDatMessage(currentRequestID, currentSubRequestID, currentBlockID, missingChunks)
			} else {
				// no missing chunks, send ACK
				sendAckMessage(currentRequestID, currentSubRequestID)
			}
		case core.PacketTypeReq_PSH_EOF:
			// Handle PSH-EOF
			currentFile.WriteToFile()
			sendAckMessage(currentRequestID, 0)
		}
	}
}
