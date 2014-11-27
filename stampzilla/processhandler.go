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

type processHandler struct {
	Processes []*Process
}

func (t *processHandler) Install(c *cli.Context) {
	requireRoot()

	i := &Installer{}

	if c.Bool("u") {
		fmt.Println("Updating stampzilla")
	} else {
		fmt.Println("Installing stampzilla")

		// Create required user and folders
		i.createUser("stampzilla")
		i.createDirAsUser("/var/spool/stampzilla", "stampzilla")
		i.createDirAsUser("/var/spool/stampzilla/config", "stampzilla")
		i.createDirAsUser("/var/log/stampzilla", "stampzilla")
		i.createDirAsUser("/home/stampzilla/go", "stampzilla")
		i.config()
	}

	//Install all
	i.goGet("github.com/stampzilla/stampzilla-go/stampzilla-server", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-chromecast", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-enocean", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-lifx", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-simple", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-stamp-amber-lights", c.Bool("u"))
	i.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-telldus", c.Bool("u"))
	//t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-telldus-events", c.Bool("u")) // #include <telldus-core.h> ERROR
	if c.Bool("u") { //Update only. do nothing more!
		return
	}

	i.bower()

	//TODO
	// create default /etc/stampzilla.conf if it does not exist
}

func (t *processHandler) Start(c *cli.Context) {
	requireRoot()

	what := c.Args().First()
	if what != "" {
		for _, what := range c.Args() {
			process := &Process{
				Pidfile: PidFile("/var/spool/stampzilla/" + what + ".pid"),
				Name:    "stampzilla-" + what,
				Command: "stampzilla-" + what,
			}
			process.start()
		}
		return
	}

	config := &Config{}
	config.readConfigFromFile("/etc/stampzilla.conf")
	fmt.Println(config)

	//TODO start all configured in our /etc/stampzilla.conf json file
	// example:
	/*
		{
			autostart:[
				{
					name: "simple",
					config: "/path/to/config/dir", //default to /var/spool/stampzilla/config/stampzilla-xxxx
				},
			]
		}
	*/

}

func (t *processHandler) Stop(c *cli.Context) {
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
	}

	//stop all running stampzilla processes
	processes := t.getRunningProcesses()
	for _, p := range processes {
		p.stop()
	}
}
func (t *processHandler) Status(c *cli.Context) {
	processes := t.getRunningProcesses()
	if len(processes) == 0 {
		fmt.Println("No stampzilla processes are running.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Name", "Pid", "CPU", "Memory")
	for _, p := range processes {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", p.Name, p.Pid, p.Status.CPU, p.Status.Memory)
	}

	w.Flush()
}

func (t *processHandler) Debug(c *cli.Context) {
	requireRoot()

	what := c.Args().First()
	//out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	toRun := "$GOPATH/bin/stampzilla-" + what
	cmd := exec.Command("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	//out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/stampzilla-server/public",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (t *processHandler) getRunningProcesses() []*Process {
	var processes []*Process

	pidFiles, err := ioutil.ReadDir("/var/spool/stampzilla")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, file := range pidFiles {
		if file.IsDir() {
			continue
		}
		pidFile := PidFile("/var/spool/stampzilla/" + file.Name())
		pid := pidFile.read()
		process := &Process{Pid: pid}
		process.Pidfile = pidFile
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
			processes = processes[:index+copy(processes[index:], processes[index+1:])]
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
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/stampzilla-server/public",
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
		os.Exit(0)
	}
} // }}}
