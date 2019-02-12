package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339, FullTimestamp: true})

	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = VERSION + " (" + BUILD_DATE + ")"
	app.Usage = "Manage stampzilla on the command line"
	app.EnableBashCompletion = true

	logrus.SetLevel(logrus.InfoLevel)

	defaultInitSystem := "bare"
	_, err := exec.LookPath("systemctl")
	if err == nil {
		defaultInitSystem = "systemd"
	}

	cliHandler := &cliHandler{}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "init-system",
			Usage: "can be bare or systemd",
			Value: defaultInitSystem,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start processes",
			Action: addDebug(cliHandler.Start),
		},
		{
			Name:   "stop",
			Usage:  "stop processes",
			Action: addDebug(cliHandler.Stop),
		},
		{
			Name:      "restart",
			ShortName: "r",
			Usage:     "restart processes",
			Action:    addDebug(cliHandler.Restart),
		},
		{
			Name:      "status",
			ShortName: "st",
			Usage:     "show process status",
			Action:    addDebug(cliHandler.Status),
		},
		{
			Name:   "debug",
			Usage:  "Start one process and get stdout and stderr print on console.",
			Action: addDebug(cliHandler.Debug),
		},
		{
			Name:      "log",
			ShortName: "l",
			Usage:     "Open the log of the supplied process in less",
			Action:    addDebug(cliHandler.Log),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "follow the log in real time",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "Lists avilable releases",
			Action: addDebug(cliHandler.List),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "d",
					Usage: "Show debug output",
				},
			},
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "Downloads and installs all stampzilla nodes and the server from precompiled binaries",
			Action:    addDebug(cliHandler.Install),
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
			Name:      "build",
			ShortName: "b",
			Usage:     "Compile and install stampzilla nodes. If none is specified, all available nodes will be installed",
			Action:    addDebug(cliHandler.Build),
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

	err = app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}

}

func addDebug(in func(c *cli.Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Bool("d") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Info("Debug output activated")
		}
		return in(c)
	}
}
