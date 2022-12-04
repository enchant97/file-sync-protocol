package main

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	PacketTypeSYN uint8 = 1
	PacketTypeACK uint8 = 2
)

func makeMessageSection(rawSection []byte) []byte {
	rawSectionLen := make([]byte, 8)
	binary.BigEndian.PutUint64(rawSectionLen, uint64(len(rawSection)))

	return append(rawSectionLen, rawSection...)
}

func makeMessage(maxLength int, messageType uint8, header protoreflect.ProtoMessage, meta protoreflect.ProtoMessage, payload io.Reader) []byte {
	// make message type
	rawMessageType := make([]byte, 1)
	rawMessageType[0] = byte(messageType)

	// make header
	rawHeader, headerErr := proto.Marshal(header)
	if headerErr != nil {
		panic(headerErr)
	}
	rawHeaderSection := makeMessageSection(rawHeader)

	rawMeta, metaErr := proto.Marshal(meta)
	if metaErr != nil {
		panic(metaErr)
	}
	rawMetaSection := makeMessageSection(rawMeta)

	// calculate total length of message, +8 for payload length
	messageLength := len(rawMessageType) + len(rawHeaderSection) + len(rawMetaSection) + 8
	maxPayloadLength := maxLength - messageLength

	// read payload
	payloadLength := 0
	rawPayload := make([]byte, maxPayloadLength)
	if payload != nil {
		var payloadErr error
		payloadLength, payloadErr = payload.Read(rawPayload)
		if payloadErr != nil {
			panic(payloadErr)
		}
	}

	// NOTE this will create a lot of memory allocations
	rawMessage := make([]byte, 0)

	rawMessage = append(rawMessage, rawMessageType...)
	rawMessage = append(rawMessage, rawHeaderSection...)
	rawMessage = append(rawMessage, rawMetaSection...)
	rawPayloadLength := make([]byte, 8)
	binary.BigEndian.PutUint64(rawPayloadLength, uint64(payloadLength))
	rawMessage = append(rawMessage, rawPayloadLength...)
	rawMessage = append(rawMessage, rawPayload...)

	return rawMessage
}

func client(address string, mtu uint32) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	synMessage := makeMessage(
		int(mtu),
		PacketTypeSYN,
		&pbtypes.SynClient{
			Mtu: mtu,
		},
		nil,
		nil,
	)

	if _, err := conn.Write(synMessage); err != nil {
		panic(err)
	}
}
