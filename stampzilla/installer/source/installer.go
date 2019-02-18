package source

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
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
		return GoGet("github.com/stampzilla/stampzilla-go/stampzilla", true)
	}

	return nil
}
func (t *Installer) Install(nodes ...string) error {
	return build(nodes, false)
}
func (t *Installer) Update(nodes ...string) error {
	return build(nodes, true)
}

func build(n []string, upgrade bool) error {
	// Install only specified nodes
	for _, name := range n {
		if !strings.HasPrefix(name, "stampzilla-") {
			name = "stampzilla-" + name
		}
		err := GoGet("github.com/stampzilla/stampzilla-go/nodes/"+name, upgrade)
		if err != nil {
			return err
		}
	}

	if len(n) != 0 {
		return nil
	}

	// Install all nodes
	nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if !strings.Contains(node.Name(), "stampzilla-") {
			continue
		}

		//Skip telldus since it contains C bindings if we dont explicly requests it to install
		if node.Name() == "stampzilla-telldus" {
			continue
		}

		err := GoGet("github.com/stampzilla/stampzilla-go/nodes/"+node.Name(), upgrade)
		if err != nil {
			return err
		}
	}
	return nil
}

func GoGet(url string, update bool) error {
	var out string
	logrus.Info("go get " + filepath.Base(url) + "... ")

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir("/tmp")

	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("LookPath Error: %s", err.Error())
	}

	// If we already is stampzilla user no need to sudo!
	if user, iErr := user.Current(); iErr == nil && user.Username == "stampzilla" {
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
		logrus.Error(out)
		return err
	}

	if out != "" {
		fmt.Println(out)
	}
	return nil
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
