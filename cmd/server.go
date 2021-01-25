package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/Coderlane/minecraft-sidecart/daemon"
)

var serverAddCommand = &cli.Command{
	Name: "add",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the server",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "path",
			Usage:    "The path to the root of the server",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		client, err := NewClient()
		if err != nil {
			daemonWarning(c.App.Writer, c.App.Name)
			return err
		}
		spec := daemon.ServerSpec{
			Path: c.String("path"),
			Name: c.String("name"),
		}
		var id string
		return client.Call("Daemon.AddServer", spec, &id)
	},
}

var serverCommand = &cli.Command{
	Name:        "server",
	Subcommands: []*cli.Command{serverAddCommand},
}
