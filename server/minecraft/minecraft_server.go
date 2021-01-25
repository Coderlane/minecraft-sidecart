package minecraft

import (
	"fmt"

	config "github.com/Coderlane/go-minecraft-config"
	"github.com/Coderlane/go-minecraft-ping/mcclient"
)

// ClientBuilder creates new minecraft clients from a config
type ClientBuilder func(*config.Config) (mcclient.MinecraftClient, error)

type Server struct {
	serverDir     string
	cfg           *config.Config
	clientBuilder ClientBuilder
}

// PlayerInfo represents a minecraft player
type PlayerInfo struct {
	Name string `json:"name" firestore:"name"`
	UUID string `json:"uuid" firestore:"uuid"`
}

// ServerInfo provides information about a minecraft server
type ServerInfo struct {
	MotD   string `json:"motd" firestore:"motd"`
	Online bool   `json:"online" firestore:"online"`

	Version       string       `json:"version" firestore:"version"`
	Icon          string       `json:"icon" firestore:"icon"`
	MaxPlayers    int          `json:"max_players" firestore:"max_players"`
	OnlinePlayers int          `json:"online_players" firestore:"online_players"`
	Players       []PlayerInfo `json:"players" firestore:"players"`
}

// NewServer creates a new client connection to a minecraft server
func NewServer(serverDir string) (*Server, error) {
	return newServerWithCustomClientBuider(
		serverDir, defaultClientBuilder)
}

// NewServer creates a new client connection with a custom client
// builder, this is useful for testing.
func newServerWithCustomClientBuider(
	serverDir string, clientBuilder ClientBuilder) (*Server, error) {
	cfg, err := config.LoadConfig(serverDir)
	if err != nil {
		return nil, err
	}
	return &Server{
		serverDir:     serverDir,
		cfg:           cfg,
		clientBuilder: clientBuilder,
	}, nil
}

func (srv *Server) GetServerInfo() interface{} {
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

func (srv *Server) getClient() (mcclient.MinecraftClient, error) {
	client, err := srv.clientBuilder(srv.cfg)
	if err != nil {
		return nil, err
	}
	if err = client.Handshake(mcclient.ClientStateStatus); err != nil {
		return nil, err
	}
	return client, nil
}

func defaultClientBuilder(
	cfg *config.Config) (mcclient.MinecraftClient, error) {
	return mcclient.NewMinecraftClientFromAddress(
		fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort))
}

func cfgToOfflineServerInfo(cfg *config.Config) ServerInfo {
	return ServerInfo{
		MotD:       cfg.MotD,
		MaxPlayers: cfg.MaxPlayers,
		Online:     false,
	}
}

func statusToServerInfo(status *mcclient.StatusResponse) ServerInfo {
	players := make([]PlayerInfo, len(status.Players.Users))
	for index, player := range status.Players.Users {
		players[index].UUID = player.UUID
		players[index].Name = player.Name
	}
	return ServerInfo{
		MotD:   status.Description.Text,
		Online: true,

		Version: status.Version.Name,
		Icon:    status.Favicon,

		OnlinePlayers: status.Players.Online,
		MaxPlayers:    status.Players.Max,
		Players:       players,
	}
}
