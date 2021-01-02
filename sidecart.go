package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Coderlane/minecraft-sidecart/auth"
	"github.com/Coderlane/minecraft-sidecart/client"
)

var serverPath = flag.String("server", "./",
	"Path to the root of the server.")

func main() {
	flag.Parse()

	tsp := auth.NewDefaultTokenSourceProvider()
	_, err := tsp.TokenSource()
	if err != nil {
		fmt.Printf("Failed to get token: %v\n", err)
		os.Exit(1)
	}

	_, err = client.NewClient(*serverPath)
	fmt.Println(*serverPath)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
}
