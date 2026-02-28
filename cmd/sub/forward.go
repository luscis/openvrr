package sub

import (
	"github.com/luscis/openvrr/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Forward struct {
	Cmd
}

func (u Forward) Url(prefix string) string {
	return prefix + "/api/forward"
}

func (u Forward) List(c *cli.Context) error {
	url := u.Url(c.String("url"))

	var items []schema.IPForward
	clt := u.NewHttp(c.String("token"))
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"))
}

func (u Forward) Commands(app *App) {
	app.Command(&cli.Command{
		Name:   "forward",
		Usage:  "IP Forward Route",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all ip forward route",
				Action: u.List,
			},
		},
	})
}
