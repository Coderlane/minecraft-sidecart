// Package firebase wraps the firebase Admin API for use in a client
//go:generate go run embed.go
package firebase

import (
	"context"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"github.com/Coderlane/minecraft-sidecart/firebase/internal"
)

// App represents a firebase application
type App struct {
	APIKey            string
	AuthDomain        string
	DatabaseURL       string
	ProjectID         string
	StorageBucket     string
	MessagingSenderID string
	AppID             string
	MeasurementID     string
	ClientID          string
	ClientSecret      string
}

// NewAuth creates a new authentication client based on the `APIKey`,
// `ClientID`, and `ClientSecret` in the `App`. You can specify any number
// of additional AuthOptions when creating the new client.
func (app *App) NewAuth(opts ...AuthOption) *Auth {
	auth := &Auth{
		app:         app,
		currentUser: nil,
		userCache:   &MemoryUserCache{},
	}
	for _, opt := range opts {
		opt.Apply(auth)
	}
	auth.idpConfig = internal.NewIdpConfig(app.APIKey, auth.emulatorHost)
	// Try to fetch the current user and refresh token from a cache
	if auth.userCache == nil {
		return auth
	}
	user, _ := auth.userCache.Get("default")
	auth.currentUser = user
	return auth
}

// NewFirestore creates a new Firestore client based on the `ProjectID` in the
// App. It uses `auth` for authenticating with `auth.CurrentUser()`.
func (app *App) NewFirestore(ctx context.Context, auth *Auth) (*firestore.Client, error) {
	authOpt := option.WithTokenSource(auth)
	return firestore.NewClient(ctx, app.ProjectID, authOpt)
}
