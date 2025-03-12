package extra

import (
	"github.com/sandertv/gophertunnel/minecraft"
)

type SessionInterface interface {
	Server() ServerConnInterface
	Client() *minecraft.Conn
}
