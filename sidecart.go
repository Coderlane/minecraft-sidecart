package main

import (
	"flag"
	"fmt"

	"github.com/Coderlane/minecraft-sidecart/config"
)

var configPath = flag.String("config", "./server.properties",
	"Path to the minecraft configuration file.")

func main() {
	_, err := config.ParseConfigFile(*configPath)
	if err != nil {
		fmt.Printf("Failed to parse config file: %v\n", err)
	}
}
