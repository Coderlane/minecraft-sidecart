package server

import (
	"fmt"
	"path"

	"github.com/Coderlane/go-minecraft-ping/mcclient"

	"github.com/Coderlane/minecraft-sidecart/config"
)

// MinecraftClientBuilder creates new minecraft clients from a config
type MinecraftClientBuilder func(*config.Config) (mcclient.MinecraftClient, error)

type minecraftServer struct {
	serverDir     string
	cfg           *config.Config
	clientBuilder MinecraftClientBuilder
}

// MinecraftPlayerInfo represents a minecraft player
type MinecraftPlayerInfo struct {
	Name string `json:"name" firestore:"name"`
	ID   string `json:"id" firestore:"id"`
}

// MinecraftServerInfo provides information about a minecraft server
type MinecraftServerInfo struct {
	MotD   string `json:"motd" firestore:"motd"`
	Online bool   `json:"online" firestore:"online"`

	Version       string                `json:"version" firestore:"version"`
	MaxPlayers    int                   `json:"max_players" firestore:"max_players"`
	OnlinePlayers int                   `json:"online_players" firestore:"online_players"`
	Players       []MinecraftPlayerInfo `json:"players" firestore:"players"`
}

// NewMinecraftServer creates a new client connection to a minecraft server
func NewMinecraftServer(serverDir string) (Server, error) {
	return newMinecraftServerWithCustomClientBuider(
		serverDir, defaultMinecraftClientBuilder)
}

// NewMinecraftServer creates a new client connection with a custom client
// builder, this is useful for testing.
func newMinecraftServerWithCustomClientBuider(
	serverDir string, clientBuilder MinecraftClientBuilder) (Server, error) {
	cfg, err := config.ParseConfigFile(path.Join(serverDir, "server.properties"))
	if err != nil {
		return nil, err
	}
	return &minecraftServer{
		serverDir:     serverDir,
		cfg:           cfg,
		clientBuilder: clientBuilder,
	}, nil
}

func (srv *minecraftServer) GetServerInfo() interface{} {
	client, err := srv.getClient()
	if err != nil {
		return cfgToOfflineServerInfo(srv.cfg)
	}
	status, err := client.Status()
	if err != nil {
		return cfgToOfflineServerInfo(srv.cfg)
	}
	return statusToServerInfo(status)
}

func (srv *minecraftServer) GetType() Type {
	return ServerTypeMinecraft
}

func (srv *minecraftServer) getClient() (mcclient.MinecraftClient, error) {
	client, err := srv.clientBuilder(srv.cfg)
	if err != nil {
		return nil, err
	}
	if err = client.Handshake(mcclient.ClientStateStatus); err != nil {
		return nil, err
	}
	return client, nil
}

func defaultMinecraftClientBuilder(
	cfg *config.Config) (mcclient.MinecraftClient, error) {
	return mcclient.NewMinecraftClientFromAddress(
		fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort))
}

func cfgToOfflineServerInfo(cfg *config.Config) MinecraftServerInfo {
	return MinecraftServerInfo{
		MotD:       cfg.MotD,
		MaxPlayers: cfg.MaxPlayers,
		Online:     false,
	}
}

func statusToServerInfo(status *mcclient.StatusResponse) MinecraftServerInfo {
	players := make([]MinecraftPlayerInfo, len(status.Players.Sample))
	for index, player := range status.Players.Sample {
		players[index].ID = player.ID
		players[index].Name = player.Name
	}
	return MinecraftServerInfo{
		MotD:   status.Description.Text,
		Online: true,

		Version: status.Version.Name,

		OnlinePlayers: status.Players.Online,
		MaxPlayers:    status.Players.Max,
		Players:       players,
	}
}
