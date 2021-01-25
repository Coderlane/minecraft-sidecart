package server

import (
	"github.com/Coderlane/minecraft-sidecart/server/minecraft"
)

//go:generate mockgen -destination=mock_server.go -package=server -self_package=github.com/Coderlane/minecraft-sidecart/server github.com/Coderlane/minecraft-sidecart/server Server

// ServerType represents the type of game server
type Type int

const (
	// ServerTypeUnknown is the default (unknown) server type
	ServerTypeUnknown Type = 0
	// ServerTypeMinecraft is a minecraft server
	ServerTypeMinecraft Type = 1
)

// Server provides common functions for working with a game server
type Server interface {
	GetServerInfo() interface{}
}

func GetType(srv interface{}) Type {
	switch srv.(type) {
	case minecraft.Server:
		return ServerTypeMinecraft
	default:
		return ServerTypeUnknown
	}
}

// NewServer creates a new server connection based on the configs in serverDir
func NewServer(serverDir string) (Server, error) {
	return minecraft.NewServer(serverDir)
}
