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
			Name:      "status",
			ShortName: "st",
			Usage:     "show process status",
			Action:    handler.Status,
		},
		{
			Name:   "debug",
			Usage:  "Start one process and get stdout and stderr print on console.",
			Action: handler.Debug,
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "installs all stampzilla nodes and the server.",
			Action:    handler.Install,
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
