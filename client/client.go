package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/Coderlane/minecraft-sidecart/db"
	"github.com/Coderlane/minecraft-sidecart/firebase"
	"github.com/Coderlane/minecraft-sidecart/server"
)

const (
	defaultIDFile       string        = ".sidecart_id.txt"
	defaultPollInterval time.Duration = time.Second * 5
)

// Client represents a client connection to a game server
type Client interface {
	Start(context.Context) error
}

type client struct {
	serverDir    string
	srv          server.Server
	db           db.Database
	user         *firebase.User
	pollInterval time.Duration
}

// NewClient creates a new client connection to a server
func NewClient(serverDir string,
	user *firebase.User, db db.Database) (Client, error) {
	srv, err := server.NewServer(serverDir)
	if err != nil {
		return nil, err
	}
	return newClientWithParams(serverDir, user, srv, db, defaultPollInterval), nil
}

func newClientWithParams(serverDir string, user *firebase.User,
	srv server.Server, db db.Database, pollInterval time.Duration) Client {
	return &client{
		serverDir:    serverDir,
		srv:          srv,
		db:           db,
		pollInterval: pollInterval,
		user:         user,
	}
}

func (cln *client) getOrCreateServerID(ctx context.Context) (string, error) {
	var (
		err      error
		serverID string
	)
	serverIDPath := path.Join(cln.serverDir, defaultIDFile)
	serverIDBytes, err := ioutil.ReadFile(serverIDPath)
	serverID = string(serverIDBytes)
	if os.IsNotExist(err) {
		serverID, err = cln.db.CreateServer(ctx, cln.user.UserID,
			cln.srv.GetType(), cln.srv.GetServerInfo())
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(serverIDPath, []byte(serverID), 0600)
		if err != nil {
			return "", err
		}
	}
	return serverID, nil
}

func (cln *client) Start(ctx context.Context) error {
	serverID, err := cln.getOrCreateServerID(ctx)
	if err != nil {
		return err
	}
	timer := time.NewTimer(cln.pollInterval)
	var lastServerInfo interface{}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			serverInfo := cln.srv.GetServerInfo()
			if reflect.DeepEqual(serverInfo, lastServerInfo) {
				continue
			}
			fmt.Println("Change detected: Updating server info.")
			err = cln.db.UpdateServerInfo(ctx, serverID, serverInfo)
			if err != nil {
				fmt.Printf("Failed to update server info: %v\n", err)
				continue
			}
			lastServerInfo = serverInfo
		}
	}
}
