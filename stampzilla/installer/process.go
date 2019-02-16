package installer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Process struct {
	Name    string
	Command string
	Args    []string
	Pidfile PidFile
	ConfDir string
	Pid     int
	Status  *ProcessStatus
	process *os.Process
}

type ProcessStatus struct {
	Running bool
	CPU     string
	Memory  string
}

func NewProcess(name, configDir string) *Process {
	return &Process{
		Pidfile: PidFile("/var/spool/stampzilla/" + name + ".pid"),
		Name:    name,
		Command: name,
		ConfDir: configDir,
	}
}

func (p *Process) Start() {

	if pid := p.Pidfile.Read(); pid != 0 {
		fmt.Println("Found pid file, checking if pid exists")

		// Read pid and try to find it
		process, err := os.FindProcess(int(pid))
		if err == nil {
			// Try to signal the process
			err = process.Signal(syscall.Signal(0))
			if err == nil {
				fmt.Println("Process "+p.Name+" (pid ", pid, ") already running!")
				return
			}
		}
	}

	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	nohupbin, err := exec.LookPath("nohup")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}

	chdircmd := ""
	if p.ConfDir != "" {
		CreateDirAsUser(p.ConfDir, "stampzilla")
		chdircmd = " cd " + p.ConfDir + "; "
	}

	log.Println("Starting: " + p.Command)
	cmd := chdircmd + nohupbin + " $GOPATH/bin/" + p.Command + " > /var/log/stampzilla/" + p.Command + ".log 2>&1 & echo $! > " + p.Pidfile.String()

	out, err := Run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	if err != nil {
		fmt.Println(out)
		fmt.Println(err)
	}
}

func (p *Process) Stop() {
	log.Println("Stopping:", p.Name)
	pid := p.Pidfile.Read()
	if pid == 0 {
		log.Println("pid file not found! Process not running?")
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("os.FindProcess(): ", err)
		return
	}

	err = process.Kill()
	if err != nil {
		fmt.Println("process.Kill(): ", err)
	}
	p.Pidfile.delete()
}

func Run(head string, parts ...string) (string, error) {
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
		"PATH=" + os.Getenv("PATH"),
	}
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
