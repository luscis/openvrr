package sub

import (
	"github.com/luscis/openvrr/pkg/schema"
	"github.com/urfave/cli/v2"
)

type SNAT struct {
	Cmd
}

func (u SNAT) Url(prefix string) string {
	return prefix + "/api/snat"
}

func (u SNAT) Add(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.SNAT{
		Source:   c.String("source"),
		SourceTo: c.String("source-to"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u SNAT) Remove(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.SNAT{
		Source: c.String("source"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u SNAT) List(c *cli.Context) error {
	url := u.Url(c.String("url"))

	var items []schema.SNAT
	clt := u.NewHttp(c.String("token"))
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"))
}

func (u SNAT) Commands(app *App) {
	app.Command(&cli.Command{
		Name:  "snat",
		Usage: "Source NAT",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a snat",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Required: true},
					&cli.StringFlag{Name: "source-to", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:  "remove",
				Usage: "Remove a snat",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:   "list",
				Usage:  "List all snats",
				Action: u.List,
			},
		},
	})
}

type DNAT struct {
	Cmd
}

func (u DNAT) Url(prefix string) string {
	return prefix + "/api/dnat"
}

func (u DNAT) Add(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.DNAT{
		Dest:   c.String("dest"),
		DestTo: c.String("dest-to"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u DNAT) Remove(c *cli.Context) error {
	url := u.Url(c.String("url"))

	data := &schema.DNAT{
		Dest: c.String("dest"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}

	return nil
}

func (u DNAT) List(c *cli.Context) error {
	url := u.Url(c.String("url"))

	var items []schema.DNAT
	clt := u.NewHttp(c.String("token"))
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"))
}

func (u DNAT) Commands(app *App) {
	app.Command(&cli.Command{
		Name:  "dnat",
		Usage: "Destination NAT",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a dnat",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dest", Required: true},
					&cli.StringFlag{Name: "dest-to", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:  "remove",
				Usage: "Remove a dnat",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dest", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:   "list",
				Usage:  "List all dnats",
				Action: u.List,
			},
		},
	})
}
