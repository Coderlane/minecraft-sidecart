package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Coderlane/minecraft-sidecart/auth"
	"github.com/Coderlane/minecraft-sidecart/client"
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
	ts = auth.InsecureDiskReuseTokenSource(*tokenPath, ts)

	_, err = client.NewClient(*serverPath)
	fmt.Println(*serverPath)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
}
