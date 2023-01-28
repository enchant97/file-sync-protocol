package core

type PacketType uint8

// UDP + Ipv4 + Ethernet II
const Ipv4ReservedBytes = (8 + 20 + 14)

const (
	PacketTypeSYN PacketType = 1
	PacketTypeACK PacketType = 2
	PacketTypeREQ PacketType = 3
	PacketTypePSH PacketType = 4
	PacketTypeFIN PacketType = 254
)
