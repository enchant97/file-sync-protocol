package core

type PacketType uint8

// UDP + Ipv4 + Ethernet II
const Ipv4ReservedBytes = (8 + 20 + 14)

const (
	PacketTypeReq_SYN     PacketType = 1
	PacketTypeReq_FIN     PacketType = 2
	PacketTypeReq_PSH     PacketType = 10
	PacketTypeReq_PSH_DAT PacketType = 11
	PacketTypeReq_PSH_VAL PacketType = 12
	PacketTypeReq_PSH_EOF PacketType = 13

	PacketTypeRes_SYN     PacketType = 1
	PacketTypeRes_ACK     PacketType = 2
	PacketTypeRes_ERR_DAT PacketType = 10
)
