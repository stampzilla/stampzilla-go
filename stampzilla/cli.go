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

	handler := &taskHandler{}

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
			Name:      "complete",
			ShortName: "c",
			Usage:     "complete a task on the list",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
	}

	app.Run(os.Args)
}
