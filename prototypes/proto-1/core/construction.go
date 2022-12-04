package core

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func makeMessageSection(rawSection []byte) []byte {
	rawSectionLen := make([]byte, 8)
	binary.BigEndian.PutUint64(rawSectionLen, uint64(len(rawSection)))

	return append(rawSectionLen, rawSection...)
}

// / Construct a message using given fields
func MakeMessage(
	maxLength int,
	messageType PacketType,
	header protoreflect.ProtoMessage,
	meta protoreflect.ProtoMessage,
	payload io.Reader,
) ([]byte, int) {
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
	var rawPayload []byte
	if payload != nil {
		rawPayload = make([]byte, maxPayloadLength)
		var payloadErr error
		payloadLength, payloadErr = payload.Read(rawPayload)
		if payloadErr != nil {
			panic(payloadErr)
		}
		// cuts out any unused (payload shorter than reserved)
		rawPayload = rawPayload[0:payloadLength]
	}

	// NOTE this will create a lot of memory allocations
	rawMessage := make([]byte, 0)

	rawMessage = append(rawMessage, rawMessageType...)
	rawMessage = append(rawMessage, rawHeaderSection...)
	rawMessage = append(rawMessage, rawMetaSection...)
	rawPayloadLength := make([]byte, 8)
	binary.BigEndian.PutUint64(rawPayloadLength, uint64(payloadLength))
	rawMessage = append(rawMessage, rawPayloadLength...)
	if rawPayload != nil {
		rawMessage = append(rawMessage, rawPayload...)
	}

	return rawMessage, payloadLength
}
