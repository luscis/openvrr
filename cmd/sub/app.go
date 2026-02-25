package sub

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	TokenFile = "/etc/openvrr/token"
)

var (
	Url     = "http://localhost:10001"
	Token   = ""
	Verbose = false
)

type App struct {
	cli    *cli.App
	Before func(c *cli.Context) error
	After  func(c *cli.Context) error
}

func (a *App) Flags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags,
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "output format: json|yaml",
			Value:   "table",
		})
	flags = append(flags,
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "admin token",
			Value:   Token,
		})
	flags = append(flags,
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"l"},
			Usage:   "api url",
			Value:   Url,
		})
	flags = append(flags,
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "enable verbose",
			Value:   false,
		})

	return flags
}

func (a *App) Init() *cli.App {
	app := &cli.App{
		Usage:    "OpenVRR utilities",
		Flags:    a.Flags(),
		Commands: []*cli.Command{},
		Before: func(c *cli.Context) error {
			if c.Bool("verbose") {
				Verbose = true
			} else {
				Verbose = false
			}
			if a.Before == nil {
				return nil
			}
			return a.Before(c)
		},
		After: func(c *cli.Context) error {
			if a.After == nil {
				return nil
			}
			return a.After(c)
		},
	}
	a.cli = app
	return a.cli
}

func (a *App) Command(cmd *cli.Command) {
	a.cli.Commands = append(a.cli.Commands, cmd)
}

func (a *App) Run(args []string) error {
	return a.cli.Run(args)
}

func Before(c *cli.Context) error {
	token := c.String("token")
	if token == "" {
		if data, err := os.ReadFile(TokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
		_ = c.Set("token", token)
	}
	return nil
}

func After(c *cli.Context) error {
	return nil
}

func Register() *App {
	app := &App{
		After:  After,
		Before: Before,
	}
	app.Init()

	Interface{}.Commands(app)
	Forward{}.Commands(app)
	SNAT{}.Commands(app)
	DNAT{}.Commands(app)

	return app
}
