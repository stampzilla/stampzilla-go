package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/facebookgo/inject"
)

func main() {
	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = "0.0.1"
	app.Usage = "Manage stampzilla on the command line"

	cliHandler := &cliHandler{}
	err := inject.Populate(cliHandler)
	if err != nil {
		panic(err)
	}

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start processes",
			Action: cliHandler.Start,
		},
		{
			Name:   "stop",
			Usage:  "start processes",
			Action: cliHandler.Stop,
		},
		{
			Name:      "restart",
			ShortName: "r",
			Usage:     "restart processes",
			Action:    cliHandler.Restart,
		},
		{
			Name:      "status",
			ShortName: "st",
			Usage:     "show process status",
			Action:    cliHandler.Status,
		},
		{
			Name:   "debug",
			Usage:  "Start one process and get stdout and stderr print on console.",
			Action: cliHandler.Debug,
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "installs all stampzilla nodes and the server.",
			Action:    cliHandler.Install,
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
