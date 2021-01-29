package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

var authSignInCommand = &cli.Command{
	Name:    "signin",
	Aliases: []string{"login"},
	Action: func(c *cli.Context) error {
		auth := c.App.Metadata["auth"].(*firebase.Auth)
		cfg := c.App.Metadata["oauth"].(*oauth2.Config)

		user, _, err := auth.SignInWithConsoleWithIO(
			c.Context, cfg, c.App.Reader, c.App.Writer)
		if err != nil {

			return err
		}
		fmt.Fprintln(c.App.Writer, "Authenticated!")

		client, err := NewClient()
		if err != nil {
			// It is OK if we can not connect to the daemon right now, we will
			// have already cached the credential. The daemon will get it the
			// next time it runs.
			daemonWarning(c.App.Writer, c.App.Name)
			return nil
		}

		var token oauth2.Token
		return client.Call("Daemon.SignIn", user, &token)
	},
}

var authSignOutCommand = &cli.Command{
	Name:    "signout",
	Aliases: []string{"logout"},
	Action: func(c *cli.Context) error {
		auth := c.App.Metadata["auth"].(*firebase.Auth)
		auth.SignOut()
		return nil
	},
}

var authCommand = &cli.Command{
	Name:        "auth",
	Subcommands: []*cli.Command{authSignInCommand, authSignOutCommand},
}
