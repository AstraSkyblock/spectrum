package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type AntiCheatInfo struct {
	Identity     string
	ClientData   []byte
	IdentityData []byte
	GameData     []byte
}

// ID ...
func (pk *AntiCheatInfo) ID() uint32 {
	return IDAntiCheatInfo
}

// Marshal ...
func (pk *AntiCheatInfo) Marshal(io protocol.IO) {
	io.String(&pk.Identity)
	io.ByteSlice(&pk.ClientData)
	io.ByteSlice(&pk.IdentityData)
	io.ByteSlice(&pk.GameData)
}
