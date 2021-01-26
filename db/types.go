package db

import (
	"github.com/Coderlane/minecraft-sidecart/server"
)

type serverDoc struct {
	Name   string      `firestore:"name"`
	Type   server.Type `firestore:"type"`
	Owners []string    `firestore:"owners"`
	Info   interface{} `firestore:"info"`
}
