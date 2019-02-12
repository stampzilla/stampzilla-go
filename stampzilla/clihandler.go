package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
	"github.com/stampzilla/stampzilla-go/stampzilla/runner"
	"github.com/urfave/cli"
)

type cliHandler struct {
}

func (t *cliHandler) UpdateConfig(c *cli.Context) error {
	requireRoot()

	oldConfig := &installer.Config{}
	oldConfig.ReadConfigFromFile("/etc/stampzilla/nodes.conf")

	newConfig := &installer.Config{}
	newConfig.GenerateDefault()
	for _, n := range newConfig.Daemons {
		if old := oldConfig.GetConfigForNode(n.Name); old != nil {
			n.Autostart = old.Autostart
		}
	}
	return newConfig.SaveToFile("/etc/stampzilla/nodes.conf")
}

func (t *cliHandler) Install(c *cli.Context) error {
	i, err := installer.New(installer.Binaries)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create installer")
		return nil
	}

	return t.runInstaller(c, i, c.Bool("u"))
}

func (t *cliHandler) Build(c *cli.Context) error {
	i, err := installer.New(installer.SourceCode)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create installer")
		return nil
	}

	return t.runInstaller(c, i, c.Bool("u"))
}

func (t *cliHandler) List(c *cli.Context) error {
	client := github.NewClient(nil)
	ctx := context.Background()
	releases, _, err := client.Repositories.ListReleases(ctx, "stampzilla", "stampzilla-go", &github.ListOptions{})

	if err != nil {
		return err
	}

	for _, v := range releases {
		fmt.Println(*v.TagName)
	}
	return nil
}

func (t *cliHandler) runInstaller(c *cli.Context, i installer.Installer, upgrade bool) error {
	requireRoot()

	err := installer.Prepare()
	if err != nil {
		return fmt.Errorf("Failed to run common prepare: %s", err)
	}

	err = i.Prepare()
	if err != nil {
		return fmt.Errorf("Failed to run installer prepare: %s", err)
	}

	if c.Bool("u") {
		err = i.Update(c.Args()...)
	} else {
		err = i.Install(c.Args()...)
	}
	if err != nil {
		return err
	}

	is := c.GlobalString("init-system")

	if is == "systemd" {
		r := &runner.Systemd{}
		if len(c.Args()) > 0 {
			for _, node := range c.Args() {
				err := r.GenerateUnit(node)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// generate for all nodes in binary dir
		nodes, err := ioutil.ReadDir("/home/stampzilla/go/bin")
		if err != nil {
			return err
		}
		for _, node := range nodes {
			if node.IsDir() {
				continue
			}
			if !strings.HasPrefix(node.Name(), "stampzilla-") {
				continue
			}
			err := r.GenerateUnit(node.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *cliHandler) Start(c *cli.Context) error {
	requireRoot()
	r := getRunner(c)
	return r.Start(c.Args()...)
}

func (t *cliHandler) Stop(c *cli.Context) error {
	requireRoot()
	r := getRunner(c)
	return r.Stop(c.Args()...)
}

func (t *cliHandler) Restart(c *cli.Context) error {
	t.Stop(c)
	t.Start(c)
	return nil
}
func (t *cliHandler) Status(c *cli.Context) error {
	r := getRunner(c)
	return r.Status()
}

func (t *cliHandler) Debug(c *cli.Context) error {
	requireRoot()

	what := c.Args().First()
	shbin, err := exec.LookPath("sh")
	if err != nil {
		return fmt.Errorf("LookPath Error: %s", err)
	}

	confDir := "/etc/stampzilla/nodes/" + what
	installer.CreateDirAsUser(confDir, "stampzilla")
	chdircmd := " cd " + confDir + "; "

	toRun := chdircmd + "$GOPATH/bin/" + runner.GetProcessName(what)
	cmd := exec.Command("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (t *cliHandler) Log(c *cli.Context) error {
	follow := c.Bool("f")
	systemd := c.GlobalString("init-system") == "systemd"

	var cmd *exec.Cmd
	if systemd {
		cmd = exec.Command("journalctl", "-u", runner.GetProcessName(c.Args().First()))
		if follow {
			cmd = exec.Command("journalctl", "-f", "-u", runner.GetProcessName(c.Args().First()))
		}
	} else {
		cmd = exec.Command("less", "-R", "/var/log/stampzilla/"+runner.GetProcessName(c.Args().First())+".log")
		if follow {
			cmd = exec.Command("tail", "-f", "/var/log/stampzilla/"+runner.GetProcessName(c.Args().First())+".log")
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getRunner(c *cli.Context) runner.Runner {
	t := c.GlobalString("init-system")
	switch t {
	case "bare":
		return &runner.Bare{}
	case "systemd":
		return &runner.Systemd{}

	}
	return nil
}

func requireRoot() { // {{{
	if os.Getuid() != 0 {
		fmt.Println("You must be root to run this")
		os.Exit(1)
	}
} // }}}
