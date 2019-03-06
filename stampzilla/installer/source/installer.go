package source

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Installer struct {
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (t *Installer) Prepare() error {
	return nil
}
func (t *Installer) Install(nodes ...string) error {
	return build(nodes, false)
}
func (t *Installer) Update(nodes ...string) error {
	return build(nodes, true)
}

func build(n []string, update bool) error {
	if len(n) == 0 {
		return fmt.Errorf("you must specify which nodes to build")
	}
	// Install only specified nodes
	for _, name := range n {
		if !strings.HasPrefix(name, "stampzilla-") {
			name = "stampzilla-" + name
		}

		if !update {
			if _, err := os.Stat(filepath.Join("/home", "stampzilla", "go", "bin", name)); err == nil {
				return fmt.Errorf("%s already installed. use -u to update", name)
			}
		}

		err := GoGet("github.com/stampzilla/stampzilla-go/nodes/" + name)
		if err != nil {
			return err
		}
	}

	return nil
}

func GoGet(url string) error {
	logrus.Info("building " + filepath.Base(url) + "... ")

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir("/tmp")

	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("LookPath Error: %s", err.Error())
	}

	_, err = Run(gobin, "get", "-u", "-d", url)
	if err != nil {
		return err
	}
	_, err = Run(gobin, getArgs(filepath.Base(url))...)
	return err

}

func getArgs(binName string) []string {
	hash, err := Run("git", "--git-dir", filepath.Join("/home", "stampzilla", "go", "src", "github.com", "stampzilla", "stampzilla-go", ".git"), "rev-parse", "--verify", "HEAD")
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("TRAVIS_COMMIT", hash)

	m := []string{
		"build",
		"-ldflags",
		"-X github.com/stampzilla/stampzilla-go/pkg/build.Version=" + os.Getenv("TRAVIS_TAG") + ` -X "github.com/stampzilla/stampzilla-go/pkg/build.BuildTime=` + time.Now().Format(time.RFC3339) + `" -X github.com/stampzilla/stampzilla-go/pkg/build.Commit=` + os.Getenv("TRAVIS_COMMIT"),
		"-o",
		filepath.Join("/home", "stampzilla", "go", "bin", binName),
		filepath.Join("/home", "stampzilla", "go", "src", "github.com", "stampzilla", "stampzilla-go", "nodes", binName),
	}
	//fmt.Println(strings.Join(m, "\n"))
	return m
}

func Run(head string, parts ...string) (string, error) {
	var err error
	var out []byte

	head, err = exec.LookPath(head)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(head, parts...)

	if current, iErr := user.Current(); iErr == nil && current.Username != "stampzilla" {
		user, err := user.Lookup("stampzilla")
		if err != nil {
			return "", err
		}
		uid, _ := strconv.Atoi(user.Uid)
		gid, _ := strconv.Atoi(user.Gid)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	cmd.Env = []string{
		"GOPATH=/home/stampzilla/go",
		"HOME=/home/stampzilla",
		"PATH=" + os.Getenv("PATH"),
	}
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error: %s : %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
