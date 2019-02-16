package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

type Bare struct {
}

func (t *Bare) Close() {
}
func (t *Bare) Start(nodes ...string) error {

	cfg := installer.Config{}
	cfg.ReadConfigFromFile("/etc/stampzilla/nodes.conf")
	installer.CreateDirAsUser("/var/log/stampzilla", "stampzilla")

	if len(nodes) > 0 {
		for _, name := range nodes {
			cfg.Start(GetProcessName(name))
		}
		return nil
	}

	for _, d := range cfg.GetAutostartingNodes() {
		cfg.Start(GetProcessName(d.Name))
	}
	return nil
}

func (t *Bare) Stop(nodes ...string) error {

	if len(nodes) > 0 {
		for _, node := range nodes {
			process := installer.NewProcess(node, "")
			process.Stop()
		}
		return nil
	}

	//stop all running stampzilla processes
	processes := t.getRunningProcesses()
	for _, p := range processes {
		p.Stop()
	}
	return nil
}

func (t *Bare) Status() error {
	processes := t.getRunningProcesses()
	if len(processes) == 0 {
		fmt.Println("No stampzilla processes are running.")
		return nil
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

	return w.Flush()
}

func (b *Bare) getRunningProcesses() []*installer.Process {
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

func GetProcessName(s string) string {
	if !strings.HasPrefix(s, "stampzilla-") {
		return "stampzilla-" + s
	}
	return s
}
