package server

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/golang/mock/gomock"

	config "github.com/Coderlane/go-minecraft-config"
	"github.com/Coderlane/go-minecraft-ping/mcclient"
)

const testServerConfig = `
motd=A Minecraft Server
rcon.port=25575
max-players=25
enable-rcon=true
rcon.password=hunter2
server-ip=
server-port=25565`

type testContext struct {
	ctrl   *gomock.Controller
	client *mcclient.MockMinecraftClient
	server Server
}

func newTestContext(t *testing.T) *testContext {
	tempDir := t.TempDir()
	testPath := path.Join(tempDir, "/server.properties")
	err := ioutil.WriteFile(testPath, []byte(testServerConfig), 0600)
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	client := mcclient.NewMockMinecraftClient(ctrl)
	server, err := newMinecraftServerWithCustomClientBuider(tempDir,
		func(*config.Config) (mcclient.MinecraftClient, error) {
			client.EXPECT().Handshake(gomock.Any()).Return(nil)
			return client, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	return &testContext{
		ctrl:   ctrl,
		server: server,
		client: client,
	}
}

func (tc *testContext) Finish() {
	tc.ctrl.Finish()
}

func TestGetMinecraftServerType(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	serverType := tc.server.GetType()
	if serverType != ServerTypeMinecraft {
		t.Errorf("Expected: %v Got: %v\n", ServerTypeMinecraft, serverType)
	}
}

func TestNewServerBadConfig(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err == nil {
		t.Errorf("Expected to fail to create server")
	}
	if server != nil {
		t.Errorf("Expected nil server")
	}
}

func TestGetMinecraftServerInfoOnline(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	status := mcclient.StatusResponse{
		Players: mcclient.StatusPlayers{
			Max:    10,
			Online: 1,
			Users: []mcclient.User{
				{"test", "ffefd5"},
			},
		},
	}
	tc.client.EXPECT().Status().Return(&status, nil)

	serverInfo := tc.server.GetServerInfo().(MinecraftServerInfo)
	if !serverInfo.Online {
		t.Errorf("Expected server to be online.")
	}
	if status.Players.Users[0].Name != serverInfo.Players[0].Name {
		t.Errorf("Expected: %s Got: %s\n",
			status.Players.Users[0].Name, serverInfo.Players[0].Name)
	}
}

func TestGetMinecraftServerInfoHandlesOffline(t *testing.T) {
	tempDir := t.TempDir()
	testPath := path.Join(tempDir, "/server.properties")
	err := ioutil.WriteFile(testPath, []byte(testServerConfig), 0600)
	if err != nil {
		t.Fatal(err)
	}

	server, err := NewServer(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	serverInfo := server.GetServerInfo().(MinecraftServerInfo)
	if serverInfo.Online {
		t.Errorf("Expected server to be offline.")
	}
}

func TestGetMinecraftServerInfoHandlesError(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	tc.client.EXPECT().Status().Return(nil, fmt.Errorf("failed to get status"))

	serverInfo := tc.server.GetServerInfo().(MinecraftServerInfo)
	if serverInfo.Online {
		t.Errorf("Expected server to be offline.")
	}
}
