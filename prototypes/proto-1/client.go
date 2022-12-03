package main

import (
	"encoding/binary"
	"net"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	PacketTypeSYN uint8 = 1
	PacketTypeACK uint8 = 2
)

func makeMessage(messageType uint8, header protoreflect.ProtoMessage) []byte {
	rawPacketDescriptor := make([]byte, 9)
	rawPacketDescriptor[0] = byte(messageType)

	rawHeader, _ := proto.Marshal(header)

	headerLen := len(rawHeader)
	binary.LittleEndian.PutUint64(rawPacketDescriptor[1:], uint64(headerLen))

	return append(rawPacketDescriptor, rawHeader...)
}

func client(address string, mtu uint32) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	synMessage := makeMessage(
		PacketTypeSYN,
		&pbtypes.SynClient{
			Mtu: mtu,
		},
	)

	if _, err := conn.Write(synMessage); err != nil {
		panic(err)
	}
}
