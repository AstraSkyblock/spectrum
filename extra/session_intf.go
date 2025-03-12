package extra

import (
	"github.com/AstraSkyblock/spectrum/server"
	"github.com/sandertv/gophertunnel/minecraft"
)

type SessionInterface interface {
	Server() *server.Conn
	Client() *minecraft.Conn
}
