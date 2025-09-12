package sub

import (
	"github.com/urfave/cli/v2"
)

type Interface struct {
}

func (u Interface) Url(prefix, name string) string {
	return ""
}

func (u Interface) Commands(app *App) {
	app.Command(&cli.Command{
		Name:  "interface",
		Usage: "Configure interfaces",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a virtual interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
			},
			{
				Name:  "remove",
				Usage: "Remove a virtual interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
			},
			{
				Name:  "list",
				Usage: "List all interfaces",
				Flags: []cli.Flag{},
			},
			VLAN{}.Commands(),
			Address{}.Commands(),
		},
	})
}

type VLAN struct {
}

func (s VLAN) Set(c *cli.Context) error {
	return nil
}

func (s VLAN) Commands() *cli.Command {
	return &cli.Command{
		Name:  "vlan",
		Usage: "Configure VLAN",
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add a vlan",
				Action: s.Set,
			},
			{
				Name:   "remove",
				Usage:  "Remove a vlan",
				Action: s.Set,
			},
		},
	}
}

type Address struct {
}

func (s Address) Set(c *cli.Context) error {
	return nil
}

func (s Address) Commands() *cli.Command {
	return &cli.Command{
		Name:  "address",
		Usage: "Configure adress",
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add a address",
				Action: s.Set,
			},
			{
				Name:   "remove",
				Usage:  "Remove a Address",
				Action: s.Set,
			},
		},
	}
}
