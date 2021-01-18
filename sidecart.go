package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Coderlane/minecraft-sidecart/client"
	"github.com/Coderlane/minecraft-sidecart/db"
	"github.com/Coderlane/minecraft-sidecart/firebase"
)

var serverPath = flag.String("server", "./",
	"Path to the root of the server.")

var tokenPath = flag.String("token", "./refresh_token.json",
	"Path to the refresh token.")

func main() {
	flag.Parse()

	ctx := context.Background()
	app := firebase.DefaultApp
	uc := NewLocalUserCache(app.ProjectID)
	auth := app.NewAuth(firebase.WithUserCache(uc))
	if auth.CurrentUser() == nil {
		cfg, err := firebase.NewOAuthConfig(firebase.GoogleAuthProvider)
		if err != nil {
			fmt.Printf("Failed to setup auth provider: %v\n", err)
			os.Exit(1)
		}
		_, _, err = auth.SignInWithConsole(ctx, cfg)
		if err != nil {
			fmt.Printf("Failed to authenticate: %v\n", err)
			os.Exit(1)
		}
	}

	firestore, err := app.NewFirestore(ctx, auth)
	if err != nil {
		fmt.Printf("Failed to initialize firestore: %v\n", err)
		os.Exit(1)
	}

	database, err := db.NewDatabase(ctx, firestore)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	cln, err := client.NewClient(*serverPath, auth.CurrentUser(), database)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	err = cln.Start(ctx)
	if err != nil {
		fmt.Printf("Failed to start client: %v\n", err)
		os.Exit(1)
	}
}
