package main

import (
	"fmt"
	"os"

	"github.com/Coderlane/minecraft-sidecart/cmd"
	"github.com/Coderlane/minecraft-sidecart/firebase"

	"github.com/urfave/cli/v2"
)

func main() {
	app := firebase.DefaultApp
	uc := NewLocalUserCache(app.ProjectID)
	auth := app.NewAuth(firebase.WithUserCache(uc))

	cliApp := cli.NewApp()
	cliApp.Usage = "A helper for minecraft servers"
	cliApp.Metadata = make(map[string]interface{})
	cliApp.Metadata["app"] = app
	cliApp.Metadata["auth"] = auth
	cliApp.Metadata["oauth"] =
		firebase.NewOAuthConfig(firebase.GoogleAuthProvider)
	cliApp.Commands = cmd.Commands
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
