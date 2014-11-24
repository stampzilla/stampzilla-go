package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
)

type taskHandler struct {
}

func (t *taskHandler) Start(c *cli.Context) {
	what := c.Args().First()
	if what != "" {
		t.start(what)
	}
}

func (t *taskHandler) Stop(c *cli.Context) {
	what := c.Args().First()
	if what != "" {
		t.stop(what)
	}
}

func (t *taskHandler) stop(app string) {
	log.Println("Stopping: stampzilla-" + app)
	pid, err := ioutil.ReadFile("/var/spool/stampzilla/" + app + ".pid")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Println("pid file not found! Process not running?")
			return
		}
		fmt.Print(err)
		return
	}
	pidString := strings.Trim(string(pid), "\n")
	_, err = run("kill", pidString)
	if err != nil {
		fmt.Println(err)
		return
	}
	//remove pid file
	os.Remove("/var/spool/stampzilla/" + app + ".pid")

}

func (t *taskHandler) start(app string) {
	log.Println("Starting: stampzilla-" + app)
	cmd := "nohup stampzilla-" + app + " > /var/log/stampzilla/" + app + " 2>&1 & echo $! > /var/spool/stampzilla/" + app + ".pid"
	run("sh", "-c", cmd)
}
func run(head string, parts ...string) (string, error) { // {{{
	var err error
	var out []byte

	head, err = exec.LookPath(head)
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	out, err = exec.Command(head, parts...).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
} // }}}
