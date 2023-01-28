package core

import (
	"encoding/binary"
	"fmt"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-2b/pbtypes"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Message struct {
	MessageType PacketType
	Header      protoreflect.ProtoMessage
	Meta        protoreflect.ProtoMessage
	Payload     []byte
}

func getHeader(
	messageType PacketType,
	isClient bool,
	rawHeader []byte,
) (protoreflect.ProtoMessage, error) {
	var header protoreflect.ProtoMessage
	var err error

	if isClient {
		// message from client
		switch messageType {
		case PacketTypeSYN:
			message := pbtypes.SynClient{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		case PacketTypeREQ:
			message := pbtypes.ReqClient{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		case PacketTypePSH:
			message := pbtypes.PshClient{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		case PacketTypeFIN:
			message := pbtypes.FinClient{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		default:
			err = fmt.Errorf("invalid client message type")
		}
	} else {
		// message from server
		switch messageType {
		case PacketTypeREQ:
			message := pbtypes.ReqServer{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		case PacketTypeACK:
			message := pbtypes.AckServer{}
			err = proto.Unmarshal(rawHeader, &message)
			header = &message
		default:
			err = fmt.Errorf("invalid server message type")
		}
	}

	return header, err
}

func getMeta(
	header protoreflect.ProtoMessage,
	isClient bool,
	rawMeta []byte,
) (protoreflect.ProtoMessage, error) {
	var meta protoreflect.ProtoMessage
	var err error

	if isClient {
		// message from client
		switch header := header.(type) {
		case *pbtypes.ReqClient:
			switch header.Type {
			case pbtypes.ReqTypes_REQ_PUSH_OBJ:
				message := pbtypes.ReqPshClient{}
				err = proto.Unmarshal(rawMeta, &message)
				meta = &message
			case pbtypes.ReqTypes_REQ_PUSH_VERIFY:
				message := pbtypes.ReqPshVerifyClient{}
				err = proto.Unmarshal(rawMeta, &message)
				meta = &message
			default:
				err = fmt.Errorf("invalid REQ client meta type")
			}
		default:
			err = fmt.Errorf("client message does not support meta")
		}
	} else {
		// message from server
		switch header := header.(type) {
		case *pbtypes.AckServer:
			switch header.Type {
			case pbtypes.AckTypes_ACK_SYN:
				message := pbtypes.AckSynServer{}
				err = proto.Unmarshal(rawMeta, &message)
				meta = &message
			default:
				err = fmt.Errorf("invalid ACK server meta type")
			}
		case *pbtypes.ReqServer:
			switch header.Type {
			case pbtypes.ReqTypes_REQ_RESEND_CHUNK:
				message := pbtypes.ReqResendChunk{}
				err = proto.Unmarshal(rawMeta, &message)
				meta = &message
			default:
				err = fmt.Errorf("invalid REQ server meta type")
			}
		default:
			err = fmt.Errorf("server message does not support meta")
		}
	}

	return meta, err
}

func GetMessage(rawMessage []byte, isClient bool) Message {
	// Process message type
	offsetIndex := 0
	messageType := PacketType(rawMessage[offsetIndex])

	// Find header length
	offsetIndex += 1
	headerLength := binary.BigEndian.Uint64(rawMessage[offsetIndex:9])

	// Process header
	offsetIndex += 8
	header, _ := getHeader(messageType, isClient, rawMessage[offsetIndex:offsetIndex+int(headerLength)])

	// Find meta length
	offsetIndex += int(headerLength)
	metaLength := binary.BigEndian.Uint64(rawMessage[offsetIndex : offsetIndex+9])

	// process meta
	offsetIndex += 8
	meta, _ := getMeta(header, isClient, rawMessage[offsetIndex:offsetIndex+int(metaLength)])

	// find payload length
	offsetIndex += int(metaLength)
	payloadLength := binary.BigEndian.Uint64(rawMessage[offsetIndex : offsetIndex+8])

	// process payload
	offsetIndex += 8
	payload := rawMessage[offsetIndex : offsetIndex+int(payloadLength)]

	return Message{
		MessageType: messageType,
		Header:      header,
		Meta:        meta,
		Payload:     payload,
	}
}
