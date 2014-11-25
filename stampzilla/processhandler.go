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
	if c.Bool("u") {
		fmt.Println("Updating stampzilla")
	} else {
		fmt.Println("Installing stampzilla")
	}

	//Install all
	t.goGet("github.com/stampzilla/stampzilla-go/stampzilla-server", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-chromecast", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-enocean", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-lifx", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-simple", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-stamp-amber-lights", c.Bool("u"))
	t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-telldus", c.Bool("u"))
	//t.goGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-telldus-events", c.Bool("u")) // #include <telldus-core.h> ERROR
	if c.Bool("u") { //Update only. do nothing more!
		return
	}

	//TODO
	// create /var/spool/stampzilla/
	// create /var/log/stampzilla/
	// create stampzilla user if it does not exist
	// run with this instead: sudo -u stampzilla -H GOPATH='$HOME/go' PATH='$GOPATH/bin' sh -c stampzilla-simple
	// create default /etc/stampzilla.conf if it does not exist
	// go get all the nodes and server. if -u is present only do this with -u and then return.(DONE)

	//should do this as root and should do the install into this user?
	//t.createUser("stampzilla")

}

func (t *processHandler) createUser(username string) {
	if t.userExists(username) {
		fmt.Println("User " + username + " already exists.")
		return
	}

	out, err := run("useradd", "-m", "-r", "-s", "/bin/false", username)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(out)
	return
}

func (t *processHandler) userExists(username string) bool {
	_, err := run("id", "-u", username)
	if err != nil {
		return false
	}
	return true
}
func (t *processHandler) goGet(url string, update bool) {
	var out string
	var err error
	fmt.Print(filepath.Base(url) + "... ")

	gobin, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	if update {
		//out, err = run("go", "get", "-u", url)
		out, err = run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", "-u", url)
	} else {
		//out, err = run("sudo", "-E", "-u", "stampzilla", "-H", "/usr/bin/env")
		out, err = run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", url)
		//out, err = run("go", "get", url)
	}
	if err != nil {
		fmt.Println(err)
		fmt.Println(out)
		return
	}
	fmt.Println("DONE")
	fmt.Println(out)
}
func (t *processHandler) Start(c *cli.Context) {
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
	}

	//TODO start all configured in our /etc/stampzilla.conf json file
	// example:
	/*
		{
			autostart:[
				{
					name: "simple",
					config: "/path/to/config/dir",
				},
			]
		}
	*/

}

func (t *processHandler) Stop(c *cli.Context) {
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
	what := c.Args().First()
	cmd := exec.Command("stampzilla-" + what)
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
			process.Name = pslineslice[len(pslineslice)-1]
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
		fmt.Printf("LookPath Error: %s", err)
	}
	cmd := exec.Command(head, parts...)
	//cmd.Env = []string{"GOPATH=$HOME/go", "PATH=$PATH:$GOPATH/bin"}
	cmd.Env = []string{"GOPATH=/home/stampzilla/go"}
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
} // }}}
