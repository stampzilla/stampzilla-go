package main

import "os"

type Process struct {
	Name     string
	Command  string
	Args     []string
	Pidfile  PidFile
	Logfile  string
	Respawn  int
	Pid      int
	Status   string
	process  *os.Process
	respawns int
}
