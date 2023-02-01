package core

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func makeMessageSection(rawSection []byte) []byte {
	rawSectionLen := make([]byte, 4)
	binary.BigEndian.PutUint32(rawSectionLen, uint32(len(rawSection)))

	return append(rawSectionLen, rawSection...)
}

func PredictMessageSize(header protoreflect.ProtoMessage) int {
	// reserved bytes for message type field
	size := 1
	// reserved bytes for header length field
	size += 4
	if header != nil {
		size += proto.Size(header)
	}
	// reserved bytes for payload length field
	size += 4
	return size
}

// Construct a message using given fields,
// ensuring that the given message will fit
func MakeMessage(
	maxLength int,
	messageType PacketType,
	header protoreflect.ProtoMessage,
	payload io.Reader,
) ([]byte, int, error) {

	// check if message will fit in given max size
	predictedSize := PredictMessageSize(header)
	if predictedSize > maxLength {
		return nil, 0, fmt.Errorf(
			"message will not fit given max length (need %d, have %d)",
			predictedSize, maxLength)
	} else if predictedSize == maxLength && payload != nil {
		return nil, 0, fmt.Errorf("message leaves no space for given payload")
	}

	// make message type
	rawMessageType := make([]byte, 1)
	rawMessageType[0] = byte(messageType)

	// make header
	rawHeader, headerErr := proto.Marshal(header)
	if headerErr != nil {
		panic(headerErr)
	}
	rawHeaderSection := makeMessageSection(rawHeader)

	// calculate total length of message, +4 for payload length
	messageLength := len(rawMessageType) + len(rawHeaderSection) + 4
	maxPayloadLength := maxLength - messageLength

	// read payload
	payloadLength := 0
	var rawPayload []byte
	if payload != nil {
		rawPayload = make([]byte, maxPayloadLength)
		var payloadErr error
		payloadLength, payloadErr = payload.Read(rawPayload)
		if payloadErr != nil && payloadErr != io.EOF {
			panic(payloadErr)
		}
		// cuts out any unused (payload shorter than reserved)
		rawPayload = rawPayload[0:payloadLength]
	}

	// NOTE this will create a lot of memory allocations
	rawMessage := make([]byte, 0)

	rawMessage = append(rawMessage, rawMessageType...)
	rawMessage = append(rawMessage, rawHeaderSection...)
	rawPayloadLength := make([]byte, 4)
	binary.BigEndian.PutUint32(rawPayloadLength, uint32(payloadLength))
	rawMessage = append(rawMessage, rawPayloadLength...)
	if rawPayload != nil {
		rawMessage = append(rawMessage, rawPayload...)
	}

	return rawMessage, payloadLength, nil
}
