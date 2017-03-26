package source

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Installer struct {
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (t *Installer) Prepare() error {
	_, err := os.Stat("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if os.IsNotExist(err) {
		logrus.Info("Found no nodes. Installing stampzilla cli first!")
		GoGet("github.com/stampzilla/stampzilla-go/stampzilla", true)
	}

	return nil
}
func (t *Installer) Install(nodes ...string) {
	build(nodes, false)
}
func (t *Installer) Update(nodes ...string) {
	build(nodes, true)
}

func build(n []string, upgrade bool) {
	// Install only specified nodes
	for _, name := range n {
		node := "stampzilla-" + name
		GoGet("github.com/stampzilla/stampzilla-go/nodes/"+node, upgrade)
	}

	if len(n) != 0 {
		return
	}

	// Install all nodes
	nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, node := range nodes {
		if !strings.Contains(node.Name(), "stampzilla-") {
			continue
		}

		//Skip telldus-events since it contains C bindings if we dont explicly requests it to install
		if node.Name() == "stampzilla-telldus-events" {
			continue
		}

		GoGet("github.com/stampzilla/stampzilla-go/nodes/"+node.Name(), upgrade)
	}
}

func GoGet(url string, update bool) {
	var out string
	var err error
	logrus.Info("go get " + filepath.Base(url) + "... ")

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir("/tmp")

	gobin, err := exec.LookPath("go")
	if err != nil {
		logrus.Error("LookPath Error: %s", err)
	}

	// If we already is stampzilla user no need to sudo!
	if user, err := user.Current(); err == nil && user.Username == "stampzilla" {
		if update {
			out, err = Run(gobin, "get", "-u", url)
		} else {
			out, err = Run(gobin, "get", url)
		}
	} else {
		if update {
			out, err = Run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", "-u", url)
		} else {
			out, err = Run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", url)
		}
	}

	if err != nil {
		logrus.Error(err)
		logrus.Debug(out)
		return
	}

	if out != "" {
		fmt.Println(out)
	}
}

func Run(head string, parts ...string) (string, error) { // {{{
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
		"STAMPZILLA_WEBROOT=/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public/dist",
	}
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
} // }}}
