package db

import (
	"github.com/Coderlane/minecraft-sidecart/server"
)

type serverDoc struct {
	Type   server.Type
	Owners []string
	Info   interface{}
}
