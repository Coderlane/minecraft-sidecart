package server

import (
	"fmt"

	config "github.com/Coderlane/go-minecraft-config"
	"github.com/Coderlane/go-minecraft-ping/mcclient"
)

// MinecraftClientBuilder creates new minecraft clients from a config
type MinecraftClientBuilder func(*config.Config) (mcclient.MinecraftClient, error)

type minecraftServer struct {
	serverDir       string
	cfg             *config.Config
	operatorPlayers config.OperatorUserList
	allowPlayers    config.AllowUserList
	denyPlayers     config.DenyUserList
	denyIPs         config.DenyIPList

	clientBuilder MinecraftClientBuilder
}

// MinecraftPlayerInfo represents a minecraft player
type MinecraftPlayerInfo struct {
	Name string `json:"name" firestore:"name"`
	UUID string `json:"uuid" firestore:"uuid"`
}

// MinecraftServerInfo provides information about a minecraft server
type MinecraftServerInfo struct {
	MotD            string                  `json:"motd" firestore:"motd"`
	Icon            string                  `json:"icon" firestore:"icon"`
	MaxPlayers      int                     `json:"max_players" firestore:"max_players"`
	OnlinePlayers   int                     `json:"online_players" firestore:"online_players"`
	Players         []MinecraftPlayerInfo   `json:"players" firestore:"players"`
	OperatorPlayers config.OperatorUserList `json:"op_players" firestore:"op_players"`
	AllowPlayers    config.AllowUserList    `json:"allow_players" firestore:"allow_players"`
	DenyPlayers     config.DenyUserList     `json:"deny_players" firestore:"deny_players"`
	DenyIPs         config.DenyIPList       `json:"deny_ips" firestore:"deny_ips"`

	Online  bool   `json:"online" firestore:"online"`
	Version string `json:"version" firestore:"version"` // Online Only
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
	cfg, err := config.LoadConfig(serverDir)
	if err != nil {
		return nil, err
	}
	operatorPlayers, err := config.LoadOperatorUserList(serverDir)
	if err != nil {
		return nil, err
	}
	allowPlayers, err := config.LoadAllowUserList(serverDir)
	if err != nil {
		return nil, err
	}
	denyPlayers, err := config.LoadDenyUserList(serverDir)
	if err != nil {
		return nil, err
	}
	denyIPs, err := config.LoadDenyIPList(serverDir)
	if err != nil {
		return nil, err
	}
	return &minecraftServer{
		serverDir:       serverDir,
		cfg:             cfg,
		operatorPlayers: operatorPlayers,
		allowPlayers:    allowPlayers,
		denyPlayers:     denyPlayers,
		denyIPs:         denyIPs,
		clientBuilder:   clientBuilder,
	}, nil
}

func (srv *minecraftServer) GetServerInfo() interface{} {
	info := srv.cfgToOfflineServerInfo()
	client, err := srv.getClient()
	if err != nil {
		return info
	}
	status, err := client.Status()
	if err != nil {
		return info
	}
	addStatusToServerInfo(status, &info)
	return info
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

func (srv *minecraftServer) cfgToOfflineServerInfo() MinecraftServerInfo {
	return MinecraftServerInfo{
		MotD:       srv.cfg.MotD,
		MaxPlayers: srv.cfg.MaxPlayers,
		Online:     false,

		OperatorPlayers: srv.operatorPlayers,
		AllowPlayers:    srv.allowPlayers,
		DenyPlayers:     srv.denyPlayers,
		DenyIPs:         srv.denyIPs,
	}
}

func defaultMinecraftClientBuilder(
	cfg *config.Config) (mcclient.MinecraftClient, error) {
	return mcclient.NewMinecraftClientFromAddress(
		fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort))
}

func addStatusToServerInfo(
	status *mcclient.StatusResponse, info *MinecraftServerInfo) {
	players := make([]MinecraftPlayerInfo, len(status.Players.Users))
	for index, player := range status.Players.Users {
		players[index].UUID = player.UUID
		players[index].Name = player.Name
	}
	info.MotD = status.Description.Text
	info.Online = true
	info.Version = status.Version.Name
	info.Icon = status.Favicon
	info.OnlinePlayers = status.Players.Online
	info.MaxPlayers = status.Players.Max
	info.Players = players
}
