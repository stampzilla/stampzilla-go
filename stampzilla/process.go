package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Process struct {
	Name     string
	Command  string
	Args     []string
	Pidfile  PidFile
	Logfile  string
	ConfDir  string
	Respawn  int
	Pid      int
	Status   *ProcessStatus
	process  *os.Process
	respawns int
}

type ProcessStatus struct {
	Running bool
	CPU     string
	Memory  string
}

func (p *Process) start() {

	if p.Pidfile.read() != 0 {
		fmt.Println("Process " + p.Name + " already running!")
		return
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
		i := &Installer{}
		i.createDirAsUser(p.ConfDir, "stampzilla")
		chdircmd = " cd " + p.ConfDir + "; "
	}

	log.Println("Starting: " + p.Command)
	cmd := chdircmd + nohupbin + " $GOPATH/bin/" + p.Command + " > /var/log/stampzilla/" + p.Command + " 2>&1 & echo $! > " + p.Pidfile.String()

	//run("sh", "-c", cmd)

	out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", cmd)
	if err != nil {
		fmt.Println(out)
		fmt.Println(err)
	}
}

func (p *Process) stop() {
	log.Println("Stopping:", p.Name)
	pid := p.Pidfile.read()
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
