package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

type cliHandler struct {
	Config    *Config    `inject:""`
	Installer *Installer `inject:""`
}

func (t *cliHandler) Install(c *cli.Context) {
	requireRoot()

	if c.Bool("u") {
		fmt.Println("Updating stampzilla")
	} else {
		fmt.Println("Installing stampzilla")

		// Create required user and folders
		t.Installer.createUser("stampzilla")
		t.Installer.createDirAsUser("/var/spool/stampzilla", "stampzilla")
		t.Installer.createDirAsUser("/var/spool/stampzilla/config", "stampzilla")
		t.Installer.createDirAsUser("/var/log/stampzilla", "stampzilla")
		t.Installer.createDirAsUser("/home/stampzilla/go", "stampzilla")
		t.Installer.config()
	}

	nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, node := range nodes {
		if !strings.Contains(node.Name(), "stampzilla-") {
			continue
		}

		if len(c.Args()) == 0 && node.Name() == "stampzilla-telldus-events" {
			continue
		}

		if len(c.Args()) != 0 && !t.findNodeInArgs(c, node.Name()) {
			continue
		}

		t.Installer.goGet("github.com/stampzilla/stampzilla-go/nodes/"+node.Name(), c.Bool("u"))

		if node.Name() == "stampzilla-server" && !c.Bool("u") {
			t.Installer.bower()
		}
	}

	return
}

func (t *cliHandler) findNodeInArgs(c *cli.Context, node string) bool {
	for _, requestedNode := range c.Args() {
		if node == "stampzilla-"+requestedNode {
			return true
		}
	}
	return false
}

func (t *cliHandler) Start(c *cli.Context) {
	requireRoot()

	t.Config.readConfigFromFile("/etc/stampzilla.conf")

	what := c.Args().First()
	if what != "" {
		for _, what := range c.Args() {
			t.start(what)
		}
		return
	}

	for _, d := range t.Config.GetAutostartingNodes() {
		t.start(d.Name)
	}
}
func (t *cliHandler) start(what string) {
	cdir := ""
	if dir := t.Config.GetConfigForNode(what); dir != nil {
		cdir = dir.Config
	}
	process := &Process{
		Pidfile: PidFile("/var/spool/stampzilla/" + what + ".pid"),
		Name:    "stampzilla-" + what,
		Command: "stampzilla-" + what,
		ConfDir: cdir,
	}
	process.start()
}

func (t *cliHandler) Stop(c *cli.Context) {
	requireRoot()

	what := c.Args().First()
	if what != "" {
		for _, what := range c.Args() {
			process := &Process{
				Pidfile: PidFile("/var/spool/stampzilla/" + what + ".pid"),
				Name:    "stampzilla-" + what,
				Command: "stampzilla-" + what,
			}
			process.stop()
		}
		return
	}

	//stop all running stampzilla processes
	processes := t.getRunningProcesses()
	for _, p := range processes {
		p.stop()
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
	//out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	t.Config.readConfigFromFile("/etc/stampzilla.conf")
	chdircmd := ""
	if dir := t.Config.GetConfigForNode(what); dir != nil {
		i := &Installer{}
		i.createDirAsUser(dir.Config, "stampzilla")
		chdircmd = " cd " + dir.Config + "; "
	}
	toRun := chdircmd + "$GOPATH/bin/stampzilla-" + what
	cmd := exec.Command("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	//out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (t *cliHandler) getRunningProcesses() []*Process {
	var processes []*Process

	pidFiles, err := ioutil.ReadDir("/var/spool/stampzilla")
	if err != nil {
		fmt.Println("/var/spool/stampzilla does not exist. Have you run stampzilla install ?")
		os.Exit(1)
	}

	for _, file := range pidFiles {
		if file.IsDir() {
			continue
		}
		pidFile := PidFile("/var/spool/stampzilla/" + file.Name())
		pid := pidFile.read()
		process := &Process{Pid: pid}
		process.Pidfile = pidFile
		process.Name = file.Name()
		processes = append(processes, process)
	}

	//change to this when you have time: http://linux.die.net/man/5/proc /proc/pid/stat
	ps, err := run("ps", "aux")
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
			var process *Process
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
			process.Status = &ProcessStatus{true, pslineslice[2], pslineslice[3]}
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

func run(head string, parts ...string) (string, error) { // {{{
	var err error
	var out []byte

	head, err = exec.LookPath(head)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(head, parts...)
	//cmd.Env = []string{"GOPATH=$HOME/go", "PATH=$PATH:$GOPATH/bin"}
	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public",
	}
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}                    // }}}
func requireRoot() { // {{{
	if os.Getuid() != 0 {
		fmt.Println("You must be root to run this")
		os.Exit(1)
	}
} // }}}
