package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/build"
	"github.com/urfave/cli/v2"
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
			&cli.BoolFlag{
				Name:  "systemd",
				Usage: "system has systemd",
				Value: true,
			},
		}
	} else {
		app.Flags = []cli.Flag{
			&cli.BoolFlag{
				Name:  "systemd",
				Usage: "system has systemd",
				Value: false,
			},
		}
	}
	app.Flags = append(app.Flags, &cli.BoolFlag{
		Name:  "debug",
		Usage: "Show debug output",
	})
	cliHandler := &cliHandler{}

	app.Commands = []*cli.Command{
		{
			Name:   "start",
			Usage:  "start processes",
			Action: (cliHandler.Start),
		},
		{
			Name:   "stop",
			Usage:  "stop processes",
			Action: (cliHandler.Stop),
		},
		{
			Name:    "restart",
			Aliases: []string{"r"},
			Usage:   "restart processes",
			Action:  (cliHandler.Restart),
		},
		{
			Name:   "disable",
			Usage:  "disable systemd service. Only valid if using systemd",
			Action: (cliHandler.Disable),
		},
		{
			Name:    "status",
			Aliases: []string{"st"},
			Usage:   "show process status",
			Action:  (cliHandler.Status),
		},
		{
			Name:   "debug",
			Usage:  "Start one process and get stdout and stderr print on console.",
			Action: (cliHandler.Debug),
		},
		{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "Open the log of the supplied process in less",
			Action:  (cliHandler.Log),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "f",
					Usage: "follow the log in real time",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "Lists available releases",
			Action: (cliHandler.List),
		},
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "Downloads and installs all stampzilla nodes and the server from precompiled binaries",
			Action:  (cliHandler.Install),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "u",
					Usage: "Force update of existing binaries",
				},
			},
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "Compile and install stampzilla nodes. If none is specified, all available nodes will be installed",
			Action:  (cliHandler.Build),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "u",
					Usage: "Force update of source files before compile",
				},
			},
		},
		{
			Name:    "updateconfig",
			Aliases: []string{"uc"},
			Usage:   "Generates new /etc/stampzilla/nodes.conf merging new nodes with existing config",
			Action:  cliHandler.UpdateConfig,
		},
		{
			Name:   "self-update",
			Usage:  "Update stampzilla cli",
			Action: (cliHandler.SelfUpdate),
		},
		{
			Name:   "version",
			Usage:  "print version of all installed stampzilla nodes",
			Action: (cliHandler.Version),
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Info("Debug output activated")
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}
}
