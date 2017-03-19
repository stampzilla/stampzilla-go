package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

func main() {
	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = VERSION + " (" + BUILD_DATE + ")"
	app.Usage = "Manage stampzilla on the command line"
	app.EnableBashCompletion = true

	logrus.SetLevel(logrus.InfoLevel)
	for _, v := range os.Args {
		if v == "-d" {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Info("Debug output activated")
		}
	}

	cliHandler := &cliHandler{}

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
			Usage:     "Downloads and installs all stampzilla nodes and the server from precompiled binaries",
			Action:    cliHandler.Install,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "u",
					Usage: "Force update of existing binaries",
				},
				cli.BoolFlag{
					Name:  "d",
					Usage: "Show debug output",
				},
			},
		},
		{
			Name:      "upgrade",
			ShortName: "u",
			Aliases:   []string{"update"},
			Usage:     "Upgrades currently installed nodes and the server",
			Action:    cliHandler.Upgrade,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "d",
					Usage: "Show debug output",
				},
			},
		},
		{
			Name:      "build",
			ShortName: "b",
			Usage:     "Compile and install stampzilla nodes. If none is specified, all available nodes will be installed",
			Action:    cliHandler.Build,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "u",
					Usage: "Force update of source files before compile",
				},
				cli.BoolFlag{
					Name:  "d",
					Usage: "Show debug output",
				},
			},
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
