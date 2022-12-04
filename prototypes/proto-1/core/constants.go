package core

type PacketType uint8

const (
	PacketTypeSYN PacketType = 1
	PacketTypeACK PacketType = 2
	PacketTypeREQ PacketType = 3
	PacketTypePSH PacketType = 4
	PacketTypeFIN PacketType = 254
)
