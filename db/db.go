package db

//go:generate mockgen -destination=mock_db.go -package=db -self_package=github.com/Coderlane/minecraft-sidecart/db github.com/Coderlane/minecraft-sidecart/db Database

import (
	"context"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"

	"github.com/Coderlane/minecraft-sidecart/auth"
	"github.com/Coderlane/minecraft-sidecart/server"
)

// Database wraps a firestore database connection
type Database interface {
	CreateServer(context.Context, server.Type, interface{}) (string, error)
	UpdateServerInfo(context.Context, string, interface{}) error
}

type database struct {
	store       *firestore.Client
	tokenSource oauth2.TokenSource
}

// NewDatabase creates a new firestore database connection
func NewDatabase(
	ctx context.Context, tokenSource oauth2.TokenSource) (Database, error) {
	authOpt := option.WithTokenSource(tokenSource)
	store, err := firestore.NewClient(ctx, "minecraft-sidecart", authOpt)
	if err != nil {
		return nil, err
	}
	return &database{
		store:       store,
		tokenSource: tokenSource,
	}, nil
}

func (db *database) CreateServer(ctx context.Context,
	serverType server.Type, serverInfo interface{}) (string, error) {
	token, err := db.tokenSource.Token()
	if err != nil {
		return "", err
	}

	userID := token.Extra(auth.IdpUserID).(string)
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
