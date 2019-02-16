package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var dest = "dist"

func main() {
	files, err := ioutil.ReadDir("nodes")
	if err != nil {
		log.Fatal(err)
	}

	os.Mkdir(dest, os.ModePerm)

	for _, dir := range files {
		path := filepath.Join("nodes", dir.Name())
		if !dir.IsDir() || !strings.HasPrefix(dir.Name(), "stampzilla-") {
			continue
		}
		if _, err := os.Stat(filepath.Join("nodes", dir.Name(), "RELEASE")); os.IsNotExist(err) {
			continue
		}

		log.Println("found RELEASE file", dir.Name())

		for _, goos := range []string{"linux"} {
			for _, goarch := range []string{"arm", "arm64", "amd64"} {
				binName := filepath.Base(fmt.Sprintf("%s-%s-%s", path, goos, goarch))
				fmt.Printf("building %s...", binName)
				cmd := exec.Command("go", getArgs(path, binName)...)
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, "GOOS="+goos, "GOARCH="+goarch)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Println(string(stdoutStderr))
					log.Fatal(err)
				}
				fmt.Printf("done\n")
			}
		}
	}
}

func getArgs(path, binName string) []string {
	m := []string{
		"build",
		"-o",
		filepath.Join(dest, binName),
	}
	filesToBuild, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		log.Fatal(err)
	}
	m = append(m, filesToBuild...)
	return m
}
