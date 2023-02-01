package core

import (
	"encoding/binary"
	"fmt"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-3/pbtypes"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Message struct {
	MessageType PacketType
	Header      protoreflect.ProtoMessage
	Payload     []byte
}

func UnmarshalHeader(b []byte, m protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
	err := proto.Unmarshal(b, m)
	return m, err
}

func getHeader(
	messageType PacketType,
	isRequest bool,
	rawHeader []byte,
) (protoreflect.ProtoMessage, error) {
	if isRequest {
		switch messageType {
		case PacketTypeReq_SYN:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqSyn{})
		case PacketTypeReq_FIN:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqFin{})
		case PacketTypeReq_PSH:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqPsh{})
		case PacketTypeReq_PSH_DAT:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqPshDat{})
		case PacketTypeReq_PSH_VAL:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqPshVal{})
		case PacketTypeReq_PSH_EOF:
			return UnmarshalHeader(rawHeader, &pbtypes.ReqPshEof{})
		default:
			return nil, fmt.Errorf("invalid request message type")
		}
	} else {
		switch messageType {
		case PacketTypeRes_SYN:
			return UnmarshalHeader(rawHeader, &pbtypes.ResSyn{})
		case PacketTypeRes_ACK:
			return UnmarshalHeader(rawHeader, &pbtypes.ResAck{})
		case PacketTypeRes_ERR_DAT:
			return UnmarshalHeader(rawHeader, &pbtypes.ResErrDat{})
		default:
			return nil, fmt.Errorf("invalid response message type")
		}
	}
}

func GetMessage(rawMessage []byte, isClient bool) Message {
	// Process message type
	offsetIndex := 0
	messageType := PacketType(rawMessage[offsetIndex])

	// Find header length
	offsetIndex += 1
	headerLength := binary.BigEndian.Uint64(rawMessage[offsetIndex:5])

	// Process header
	offsetIndex += 4
	header, _ := getHeader(messageType, isClient, rawMessage[offsetIndex:offsetIndex+int(headerLength)])

	// find payload length
	offsetIndex += int(headerLength)
	payloadLength := binary.BigEndian.Uint64(rawMessage[offsetIndex : offsetIndex+4])

	// process payload
	offsetIndex += 4
	payload := rawMessage[offsetIndex : offsetIndex+int(payloadLength)]

	return Message{
		MessageType: messageType,
		Header:      header,
		Payload:     payload,
	}
}
