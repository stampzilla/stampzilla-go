package source

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

type Installer struct {
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (t *Installer) Prepare() error {
	return nil
}
func (t *Installer) Install(nodes ...string) {

}
func (t *Installer) Update(nodes ...string) {

}

func (t *Installer) GoGet(url string, update bool) {
	var out string
	var err error
	fmt.Print("go get " + filepath.Base(url) + "... ")

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir("/tmp")

	gobin, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
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
		fmt.Println(err)
		fmt.Println(out)
		return
	}
	fmt.Println("DONE")
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
