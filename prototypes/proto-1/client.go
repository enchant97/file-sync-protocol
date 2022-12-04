package main

import (
	"net"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

func client(address string, mtu uint32) {
	s, _ := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	synMessage, _ := core.MakeMessage(
		int(mtu),
		core.PacketTypeSYN,
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
