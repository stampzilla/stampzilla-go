package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/pkg/build"
	"github.com/urfave/cli"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339, FullTimestamp: true})

	app := cli.NewApp()
	app.Name = "stampzilla"
	app.Version = build.String()
	app.Usage = "Manage stampzilla on the command line"
	app.EnableBashCompletion = true

	logrus.SetLevel(logrus.InfoLevel)

	if _, err := exec.LookPath("systemctl"); err == nil {
		app.Flags = []cli.Flag{
			cli.BoolTFlag{
				Name:  "systemd",
				Usage: "system has systemd",
			},
		}
	} else {
		app.Flags = []cli.Flag{
			cli.BoolFlag{
				Name:  "systemd",
				Usage: "system has systemd",
			},
		}

	}
	app.Flags = append(app.Flags, cli.BoolFlag{
		Name:  "debug",
		Usage: "Show debug output",
	})
	cliHandler := &cliHandler{}

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
			Name:   "disable",
			Usage:  "disable systemd service. Only valid if using systemd",
			Action: addDebug(cliHandler.Disable),
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
			},
		},
		{
			Name:      "updateconfig",
			ShortName: "uc",
			Usage:     "Generates new /etc/stampzilla/nodes.conf merging new nodes with existing config",
			Action:    cliHandler.UpdateConfig,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}

}

func addDebug(in func(c *cli.Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Info("Debug output activated")
		}
		return in(c)
	}
}
