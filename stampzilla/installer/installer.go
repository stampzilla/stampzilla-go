package installer

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

type Installer struct {
	Config *Config
}

func NewInstaller() *Installer {
	c := &Config{}
	return &Installer{c}
}

func (t *Installer) CreateConfig() {
	fmt.Print("Creating config /etc/stampzilla/nodes.conf... ")
	if _, err := os.Stat("/etc/stampzilla/nodes.conf"); os.IsNotExist(err) {
		config := &Config{}
		config.GenerateDefault()
		config.SaveToFile("/etc/stampzilla/nodes.conf")
		fmt.Println("DONE")
	} else {
		fmt.Println("Already exists, Skipping!")
	}
}

func (t *Installer) Bower() {
	bower, err := exec.LookPath("bower")
	if err != nil {
		fmt.Println("Missing bower executable. Install with: npm install -g bower")
		os.Exit(1)
	}

	fmt.Print("bower install in public folder... ")
	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
		return
	}

	toRun := "cd /home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public && " + bower + " install"
	out, err := Run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	if err != nil {
		fmt.Println("ERROR", err, out)
	}
	fmt.Println("DONE")
}
func (t *Installer) CreateUser(username string) {
	fmt.Print("Creating user " + username + "... ")
	if t.userExists(username) {
		fmt.Println("already exists!")
		return
	}

	out, err := Run("useradd", "-m", "-r", "-s", "/bin/false", username)
	if err != nil {
		fmt.Println("ERROR", err, out)
		return
	}
	fmt.Println("DONE")
}
func (t *Installer) userExists(username string) bool {
	_, err := Run("id", "-u", username)
	if err != nil {
		return false
	}
	return true
}

func (t *Installer) CreateDirAsUser(directory string, username string) {
	fmt.Print("Creating directory " + directory + "... ")

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			fmt.Println("ERROR", err)
			return
		}
	} else {
		fmt.Print("Already exists... Fixing permissions... ")
	}

	u, err := user.Lookup(username)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	err = os.Chown(directory, uid, gid)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	fmt.Println("DONE")
}

func (t *Installer) GoGet(url string, update bool) {
	var out string
	var err error
	fmt.Print("go get " + filepath.Base(url) + "... ")

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
