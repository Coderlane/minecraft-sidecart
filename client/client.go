package client

import (
	"fmt"
	"path"

	"github.com/Coderlane/go-minecraft-ping/mcclient"
	"github.com/Coderlane/minecraft-sidecart/config"
)

type Client interface {
}

type client struct {
	cfg    *config.Config
	client mcclient.MinecraftClient
}

func NewClient(serverDir string) (Client, error) {
	cfg, err := config.ParseConfigFile(path.Join(serverDir, "server.properties"))
	if err != nil {
		return nil, err
	}
	mcc, err := mcclient.NewMinecraftClientFromAddress(
		fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort))
	return &client{
		cfg:    cfg,
		client: mcc,
	}, nil
}
