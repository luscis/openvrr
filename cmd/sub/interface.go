package sub

import (
	"github.com/luscis/openvrr/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Interface struct {
	Cmd
}

func (u Interface) Url(prefix string) string {
	return prefix + "/api/interface"
}

func (u Interface) Add(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.Interface{
		Name: c.String("name"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u Interface) Remove(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.Interface{
		Name: c.String("name"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u Interface) List(c *cli.Context) error {
	url := u.Url(c.String("url"))

	var items []schema.Interface
	clt := u.NewHttp(c.String("token"))
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"))
}

func (u Interface) Commands(app *App) {
	app.Command(&cli.Command{
		Name:  "interface",
		Usage: "Network interface",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a virtual interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:  "remove",
				Usage: "Remove a virtual interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:   "list",
				Usage:  "List all interfaces",
				Action: u.List,
			},
			VLAN{}.Commands(),
			Address{}.Commands(),
		},
	})
}

type VLAN struct {
	Cmd
}

func (s VLAN) Url(prefix string) string {
	return prefix + "/api/vlan"
}

func (s VLAN) List(c *cli.Context) error {
	url := s.Url(c.String("url"))

	var items []schema.Interface
	clt := s.NewHttp(c.String("token"))
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return s.Out(items, c.String("format"))
}

func (s VLAN) Add(c *cli.Context) error {
	url := s.Url(c.String("url"))
	data := &schema.Interface{
		Name:   c.String("interface"),
		Tag:    c.Int("tag"),
		Trunks: c.String("trunks"),
	}

	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (s VLAN) Remove(c *cli.Context) error {
	url := s.Url(c.String("url"))
	data := &schema.Interface{
		Name:   c.String("interface"),
		Tag:    c.Int("tag"),
		Trunks: c.String("trunks"),
	}

	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (s VLAN) Commands() *cli.Command {
	return &cli.Command{
		Name:  "vlan",
		Usage: "Configure VLAN",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a vlan",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "interface", Required: true},
					&cli.IntFlag{Name: "tag"},
					&cli.StringFlag{Name: "trunks"},
				},
				Action: s.Add,
			},
			{
				Name:  "remove",
				Usage: "remove a vlan",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "interface", Required: true},
					&cli.IntFlag{Name: "tag", Value: 4095},
					&cli.StringFlag{Name: "trunks", Value: "all"},
				},
				Action: s.Remove,
			},
			{
				Name:   "list",
				Usage:  "List all vlan",
				Action: s.List,
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
