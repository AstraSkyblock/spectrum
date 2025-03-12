package extra

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

type ServerConnInterface interface {
	WritePacket(pk packet.Packet) error
}
