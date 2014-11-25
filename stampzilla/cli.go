package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = "0.0.1"
	app.Usage = "Manage stampzilla on the command line"

	handler := &processHandler{}

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start processes",
			Action: handler.Start,
		},
		{
			Name:   "stop",
			Usage:  "start processes",
			Action: handler.Stop,
		},
		{
			Name:  "status",
			Usage: "show process status",
			Action: func(c *cli.Context) {
				println("status")
			},
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "installs all stampzilla nodes and the server.",
			Action: func(c *cli.Context) {
				println("completed ", c.Args().First())
				println("u flag ", c.Bool("u"))
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "u",
					Usage: "Do upgrade",
				},
			},
		},
	}

	app.Run(os.Args)
}
