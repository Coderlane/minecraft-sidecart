package db

//go:generate mockgen -destination=mock_db.go -package=db -self_package=github.com/Coderlane/minecraft-sidecart/db github.com/Coderlane/minecraft-sidecart/db Database

import (
	"context"

	firestore "cloud.google.com/go/firestore"

	"github.com/Coderlane/minecraft-sidecart/server"
)

// Database wraps a firestore database connection
type Database interface {
	CreateServer(context.Context, string, server.Type, interface{}) (string, error)
	UpdateServerInfo(context.Context, string, interface{}) error
}

type database struct {
	store *firestore.Client
}

// NewDatabase creates a new firestore database connection
func NewDatabase(
	ctx context.Context, store *firestore.Client) (Database, error) {
	return &database{
		store: store,
	}, nil
}

func (db *database) CreateServer(ctx context.Context,
	userID string, serverType server.Type, serverInfo interface{}) (string, error) {
	serverDetails := serverDoc{
		Type:   serverType,
		Owners: []string{userID},
		Info:   serverInfo,
	}

	server, _, err := db.store.Collection("servers").Add(ctx, serverDetails)
	if err != nil {
		return "", err
	}
	return server.ID, nil
}

func (db *database) UpdateServerInfo(ctx context.Context,
	serverID string, serverInfo interface{}) error {
	_, err := db.store.Collection("servers").Doc(serverID).Update(
		ctx, []firestore.Update{
			{Path: "info", Value: serverInfo},
		})
	return err
}
