package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

func main() {
	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = VERSION + " (" + BUILD_DATE + ")"
	app.Usage = "Manage stampzilla on the command line"

	cliHandler := &cliHandler{installer.NewInstaller()}

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
			Name:      "log",
			ShortName: "l",
			Usage:     "Open the log of the supplied process in less",
			Action:    cliHandler.Log,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "follow the log in real time",
				},
			},
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
		{
			Name:      "upgrade",
			ShortName: "u",
			Aliases:   []string{"update"},
			Usage:     "upgrades currently installed nodes and the server",
			Action:    cliHandler.Upgrade,
		},
		{
			Name:      "updateconfig",
			ShortName: "uc",
			Usage:     "Generates new /etc/stampzilla/nodes.conf merging new nodes with existing config",
			Action:    cliHandler.UpdateConfig,
		},
	}

	app.Run(os.Args)
}
