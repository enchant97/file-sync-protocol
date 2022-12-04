package main

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/core"
	"github.com/enchant97/file-sync-protocol/prototypes/proto-1/pbtypes"
)

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

	n, addr, _ := conn.ReadFromUDP(buffer)
	fmt.Println(buffer)
	message := core.GetMessage(buffer[0:n], true)
	fmt.Println(message)

	if message.MessageType == core.PacketTypeSYN {
		ackMessage, _ := core.MakeMessage(
			int(mtu),
			core.PacketTypeACK,
			&pbtypes.AckServer{
				ReqId: 0,
				Type:  pbtypes.AckTypes_SYN,
			},
			&pbtypes.AckSynServer{
				ClientId: rand.Uint32(),
				Mtu:      mtu,
			},
			nil,
		)
		conn.WriteToUDP(ackMessage, addr)
	}
}
