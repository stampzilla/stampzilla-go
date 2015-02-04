package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

type Installer struct {
}

func (t *Installer) config() {
	fmt.Print("Creating config /etc/stampzilla.conf... ")
	if _, err := os.Stat("/etc/stampzilla.conf"); os.IsNotExist(err) {
		config := &Config{}
		config.generateDefault()
		config.SaveToFile("/etc/stampzilla.conf")
		fmt.Println("DONE")
	} else {
		fmt.Println("Already exists, Skipping!")
	}
}

func (t *Installer) bower() {
	if _, err := exec.LookPath("bower"); err != nil {
		fmt.Println("Missing bower executable. Install with: npm install -g bower")
		os.Exit(1)
	}

	fmt.Print("bower install in public folder... ")
	shbin, err := exec.LookPath("sh")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
		return
	}

	toRun := "cd /home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/public && bower install"
	out, err := run("sudo", "-E", "-u", "stampzilla", "-H", shbin, "-c", toRun)
	if err != nil {
		fmt.Println("ERROR", err, out)
	}
	fmt.Println("DONE")
}
func (t *Installer) createUser(username string) {
	fmt.Print("Creating user " + username + "... ")
	if t.userExists(username) {
		fmt.Println("already exists!")
		return
	}

	out, err := run("useradd", "-m", "-r", "-s", "/bin/false", username)
	if err != nil {
		fmt.Println("ERROR", err, out)
		return
	}
	fmt.Println("DONE")
}
func (t *Installer) userExists(username string) bool {
	_, err := run("id", "-u", username)
	if err != nil {
		return false
	}
	return true
}

func (t *Installer) createDirAsUser(directory string, username string) {
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

func (t *Installer) goGet(url string, update bool) {
	var out string
	var err error
	fmt.Print("go get " + filepath.Base(url) + "... ")

	gobin, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("LookPath Error: %s", err)
	}
	//shbin, err := exec.LookPath("sh")
	//if err != nil {
	//fmt.Printf("LookPath Error: %s", err)
	//return
	//}
	//out, err = run("sudo", "-E", "-u", "stampzilla", "-H", "/usr/bin/env")
	//fmt.Println(out)
	//return
	if update {
		//out, err = run("go", "get", "-u", url)
		out, err = run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", "-u", url)
	} else {
		out, err = run("sudo", "-E", "-u", "stampzilla", "-H", gobin, "get", url)
		//out, err = run("go", "get", url)
	}
	if err != nil {
		fmt.Println(err)
		fmt.Println(out)
		return
	}
	fmt.Println("DONE")
	//fmt.Println(out)
}
