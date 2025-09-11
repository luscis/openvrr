package main

import (
	"log"
	"os"

	"github.com/luscis/openvrr/cmd/sub"
)

func main() {
	app := sub.Register()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
