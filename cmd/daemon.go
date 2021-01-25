package cmd

import (
	"fmt"
	"io"

	"github.com/urfave/cli/v2"

	"github.com/Coderlane/minecraft-sidecart/daemon"
	"github.com/Coderlane/minecraft-sidecart/firebase"
)

const daemonWarningStr = `Warning: could not communicate with daemon.
consider starting it with '%s daemon'.
`

func daemonWarning(writer io.Writer, app string) {
	fmt.Fprintf(writer, daemonWarningStr, app)
}

var daemonCommand = &cli.Command{
	Name:  "daemon",
	Usage: "",
	Action: func(c *cli.Context) error {
		app := c.App.Metadata["app"].(*firebase.App)
		auth := c.App.Metadata["auth"].(*firebase.Auth)
		dae, err := daemon.NewRPCDaemon(c.Context, app, auth)
		if err != nil {
			return err
		}
		defer dae.Close()
		return dae.Run(c.Context)
	},
}
