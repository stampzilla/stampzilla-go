package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

type cliHandler struct {
}

func (t *cliHandler) UpdateConfig(c *cli.Context) {
	requireRoot()

	oldConfig := &installer.Config{}
	oldConfig.ReadConfigFromFile("/etc/stampzilla/nodes.conf")

	newConfig := &installer.Config{}
	newConfig.GenerateDefault()
	for _, n := range newConfig.Daemons {
		if old := oldConfig.GetConfigForNode(n.Name); old != nil {
			n.Autostart = old.Autostart
			n.Config = old.Config
		}
	}
	err := newConfig.SaveToFile("/etc/stampzilla/nodes.conf")
	if err != nil {
		fmt.Println(err)
	}

}

func (t *cliHandler) Install(c *cli.Context) {
	i, err := installer.New(installer.Binaries)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create installer")
		return
	}

	t.runInstaller(c, i, c.Bool("u"))
}
func (t *cliHandler) Upgrade(c *cli.Context) {
	i, err := installer.New(installer.Binaries)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create installer")
		return
	}

	t.runInstaller(c, i, true)
}

func (t *cliHandler) Build(c *cli.Context) {
	i, err := installer.New(installer.SourceCode)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create installer")
		return
	}

	t.runInstaller(c, i, c.Bool("u"))
}

func (t *cliHandler) List(c *cli.Context) {
	client := github.NewClient(nil)
	ctx := context.Background()
	releases, _, err := client.Repositories.ListReleases(ctx, "stampzilla", "stampzilla-go", &github.ListOptions{})

	if err != nil {
		logrus.Error(err)
		return
	}

	for _, v := range releases {
		fmt.Println(*v.TagName)
	}
}

func (t *cliHandler) runInstaller(c *cli.Context, i installer.Installer, upgrade bool) {
	requireRoot()

	err := installer.Prepare()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to run common prepare")
		return
	}

	err = i.Prepare()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to run installer prepare")
		return
	}

	if upgrade {
		i.Update(c.Args()...)
	} else {
		i.Install(c.Args()...)
	}
}

/*
func (t *cliHandler) build(c *cli.Context, upgrade bool) {
	nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		fmt.Println("Found no nodes. installing stampzilla cli first!")
		t.Installer.CreateUser("stampzilla")
		t.Installer.CreateDirAsUser("/home/stampzilla/go", "stampzilla")
		t.Installer.GoGet("github.com/stampzilla/stampzilla-go/stampzilla", upgrade)
	}

	if upgrade {
		fmt.Println("Updating stampzilla")
	} else {
		fmt.Println("Installing stampzilla")

		// Create required user and folders
		t.Installer.CreateUser("stampzilla")
		t.Installer.CreateDirAsUser("/var/spool/stampzilla", "stampzilla")
		//t.Installer.CreateDirAsUser("/var/spool/stampzilla/config", "stampzilla")
		t.Installer.CreateDirAsUser("/var/log/stampzilla", "stampzilla")
		t.Installer.CreateDirAsUser("/home/stampzilla/go", "stampzilla")
		t.Installer.CreateDirAsUser("/etc/stampzilla", "stampzilla")
		t.Installer.CreateDirAsUser("/etc/stampzilla/nodes", "stampzilla")
		t.Installer.CreateConfig()
	}

	nodes, err = ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(c.Args()) != 0 {
		t.installSpecificNodesFromArguments(c, upgrade)
		return
	}

	for _, node := range nodes {
		if !strings.Contains(node.Name(), "stampzilla-") {
			continue
		}

		//Skip telldus-events since it contains C bindings if we dont explicly requests it to install
		if len(c.Args()) == 0 && node.Name() == "stampzilla-telldus-events" {
			continue
		}

		t.Installer.GoGet("github.com/stampzilla/stampzilla-go/nodes/"+node.Name(), upgrade)
	}

	return
}

func (t *cliHandler) installSpecificNodesFromArguments(c *cli.Context, upgrade bool) {
	for _, name := range c.Args() {
		node := "stampzilla-" + name
		t.Installer.GoGet("github.com/stampzilla/stampzilla-go/nodes/"+node, upgrade)
	}
}
*/

func (t *cliHandler) Start(c *cli.Context) {
	requireRoot()

	cfg := installer.Config{}
	cfg.ReadConfigFromFile("/etc/stampzilla/nodes.conf")
	installer.CreateDirAsUser("/var/log/stampzilla", "stampzilla")

	if c.Args().First() != "" {
		for _, name := range c.Args() {
			cfg.Start(name)
		}
		return
	}

	for _, d := range cfg.GetAutostartingNodes() {
		cfg.Start(d.Name)
	}
}

func (t *cliHandler) Stop(c *cli.Context) {
	requireRoot()

	what := c.Args().First()
	if what != "" {
		for _, what := range c.Args() {
			process := installer.NewProcess(what, "")
			process.Stop()
		}
		return
	}

	//stop all running stampzilla processes
	processes := t.getRunningProcesses()
	for _, p := range processes {
		p.Stop()
	}
}
func (t *cliHandler) Restart(c *cli.Context) {
	t.Stop(c)
	t.Start(c)
}
func (t *cliHandler) Status(c *cli.Context) {
	processes := t.getRunningProcesses()
	if len(processes) == 0 {
		fmt.Println("No stampzilla processes are running.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Name", "Pid", "CPU", "Memory")
	for _, p := range processes {
		if p.Status != nil {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", p.Name, p.Pid, p.Status.CPU, p.Status.Memory)
			continue
		}

		fmt.Fprintf(w, "%s\t%d\t(DIED)\n", p.Name, p.Pid)
	}

	w.Flush()
}

func (t *cliHandler) Debug(c *cli.Context) {
	requireRoot()

	what := c.Args().First()
	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	cfg := installer.Config{}
	cfg.ReadConfigFromFile("/etc/stampzilla/nodes.conf")
	chdircmd := ""
	if dir := cfg.GetConfigForNode(what); dir != nil {
		//i := &Installer{}
		//i.CreateDirAsUser(dir.Config, "stampzilla")
		installer.CreateDirAsUser(dir.Config, "stampzilla")
		chdircmd = " cd " + dir.Config + "; "
	}
	toRun := chdircmd + "$GOPATH/bin/stampzilla-" + what
	cmd := exec.Command("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (t *cliHandler) Log(c *cli.Context) {
	follow := c.Bool("f")
	cmd := exec.Command("less", "-R", "/var/log/stampzilla/stampzilla-"+c.Args().First())
	if follow {
		cmd = exec.Command("tail", "-f", "/var/log/stampzilla/stampzilla-"+c.Args().First())
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (t *cliHandler) getRunningProcesses() []*installer.Process {
	var processes []*installer.Process

	pidFiles, err := ioutil.ReadDir("/var/spool/stampzilla")
	if err != nil {
		fmt.Println("/var/spool/stampzilla does not exist. Have you run stampzilla install ?")
		os.Exit(1)
	}

	for _, file := range pidFiles {
		if file.IsDir() {
			continue
		}
		process := installer.NewProcess(strings.TrimSuffix(file.Name(), ".pid"), "")
		process.Pid = process.Pidfile.Read()
		processes = append(processes, process)
	}

	//change to this when you have time: http://linux.die.net/man/5/proc /proc/pid/stat
	ps, err := installer.Run("ps", "aux")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	psLines := strings.Split(ps, "\n")
	for _, psline := range psLines {
		p := strings.Split(psline, " ")
		var pslineslice []string
		for _, name := range p {
			if name != "" {
				pslineslice = append(pslineslice, name)
			}
		}

		if len(pslineslice) > 1 {
			var process *installer.Process
			for _, p1 := range processes {
				process = nil
				//fmt.Println(p[4], p1.Pid)
				//fmt.Printf("%#v\n", p)
				if strings.Contains(pslineslice[1], strconv.Itoa(p1.Pid)) {
					process = p1
					break
				}

			}

			if process == nil {
				continue
			}
			//fmt.Println("NAME", p[len(p)-1])
			//fmt.Println("CPU", p[6])
			//fmt.Println("MEM", p[8])
			//process := &Process{Name: p[len(p)-1], Command: p[len(p)-1]}
			//fmt.Printf("%#v\n", pslineslice)
			process.Name = filepath.Base(pslineslice[len(pslineslice)-1])
			process.Command = pslineslice[len(pslineslice)-1]
			process.Status = &installer.ProcessStatus{true, pslineslice[2], pslineslice[3]}
		}
	}

	//remove not found processes from the list.
	for index, p := range processes {
		if p.Name == "" {
			if len(processes) > index+1 {
				processes = append(processes[:index], processes[index+1:]...)
			} else {
				processes = processes[:index]
			}
		}
	}
	return processes
}

func requireRoot() { // {{{
	if os.Getuid() != 0 {
		fmt.Println("You must be root to run this")
		os.Exit(1)
	}
} // }}}
