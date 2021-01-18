package client

import (
	"context"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/Coderlane/minecraft-sidecart/db"
	"github.com/Coderlane/minecraft-sidecart/firebase"
	"github.com/Coderlane/minecraft-sidecart/server"
)

const (
	testServerID string = "test_id"
)

type testContext struct {
	testDir string
	ctrl    *gomock.Controller
	server  *server.MockServer
	db      *db.MockDatabase
	client  Client
}

func newTestContext(t *testing.T) *testContext {
	testDir := t.TempDir()
	ctrl := gomock.NewController(t)
	server := server.NewMockServer(ctrl)
	db := db.NewMockDatabase(ctrl)
	user := &firebase.User{UserID: "test"}
	client := newClientWithParams(testDir, user, server, db, time.Millisecond*100)
	return &testContext{
		testDir: testDir,
		ctrl:    ctrl,
		server:  server,
		db:      db,
		client:  client,
	}
}

func (tc *testContext) Finish() {
	tc.ctrl.Finish()
}

func TestCreateNewServer(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	ctx, cancel := context.WithCancel(context.Background())

	serverType := server.ServerTypeMinecraft
	serverInfo := server.MinecraftServerInfo{}

	tc.server.EXPECT().GetType().Return(serverType)
	tc.server.EXPECT().GetServerInfo().Return(serverInfo)

	tc.db.EXPECT().CreateServer(ctx, "test", serverType, serverInfo).DoAndReturn(
		func(context.Context, string, server.Type, interface{}) (string, error) {
			cancel()
			return testServerID, nil
		})
	err := tc.client.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	serverIDBytes, err := ioutil.ReadFile(path.Join(tc.testDir, defaultIDFile))
	if err != nil {
		t.Fatal(err)
	}
	serverID := string(serverIDBytes)
	if testServerID != serverID {
		t.Errorf("Expected: %v Got: %v\n", testServerID, serverID)
	}
}

func TestUpdateServer(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Finish()

	err := ioutil.WriteFile(path.Join(tc.testDir, defaultIDFile),
		[]byte(testServerID), 0660)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	serverInfo := server.MinecraftServerInfo{}
	tc.server.EXPECT().GetServerInfo().Return(serverInfo)

	tc.db.EXPECT().UpdateServerInfo(ctx, testServerID, serverInfo).DoAndReturn(
		func(context.Context, string, interface{}) error {
			cancel()
			return nil
		})

	err = tc.client.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
