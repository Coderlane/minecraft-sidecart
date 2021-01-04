package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Coderlane/minecraft-sidecart/auth"
	"github.com/Coderlane/minecraft-sidecart/client"
	"github.com/Coderlane/minecraft-sidecart/db"
)

var serverPath = flag.String("server", "./",
	"Path to the root of the server.")

var tokenPath = flag.String("token", "./refresh_token.json",
	"Path to the refresh token.")

func main() {
	flag.Parse()

	ctx := context.Background()
	tsp := auth.NewFirebaseTokenSourceProvider()
	ts, err := tsp.TokenSource(ctx, auth.LoadDiskToken(*tokenPath))
	if err != nil {
		fmt.Printf("Failed to get token: %v\n", err)
		os.Exit(1)
	}

	database, err := db.NewDatabase(ctx,
		auth.InsecureDiskReuseTokenSource(*tokenPath, ts))
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	cln, err := client.NewClient(*serverPath, database)
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
