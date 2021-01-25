package cmd

import (
	"net"
	"net/rpc"

	"github.com/Coderlane/minecraft-sidecart/daemon"
)

// Create a new RPC client connected to the daemon
func NewClient() (*rpc.Client, error) {
	addr, err := daemon.DefaultAddress()
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}
