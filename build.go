package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var dest = "dist"

// Release is read from the RELEASE file and contains options for the build
type Release struct {
	Cgo  bool     `json:"cgo"`
	Arch []string `json:"arch"`
}

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

		release, _ := loadReleaseFromFile(filepath.Join("nodes", dir.Name(), "RELEASE"))
		build(path, release)
	}

	fmt.Println("Building stampzilla cli")
	build("stampzilla", nil)
}

func build(path string, release *Release) {
	for _, goos := range []string{"linux"} {
		arch := []string{"amd64", "arm", "arm64"}
		if release != nil && len(release.Arch) > 0 {
			arch = release.Arch
		}

		for _, goarch := range arch {
			binName := filepath.Base(fmt.Sprintf("%s-%s-%s", path, goos, goarch))
			cgo := "0"
			if release != nil && release.Cgo {
				cgo = "1"
			}

			fmt.Printf("building %s...\n", binName)
			cmd := exec.Command("go", getArgs(path, binName)...)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED="+cgo)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(string(stdoutStderr))
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("done\n")
}

func getArgs(path, binName string) []string {
	m := []string{
		"build",
		"-ldflags",
		"-X github.com/stampzilla/stampzilla-go/pkg/build.Version=" + os.Getenv("TRAVIS_TAG") + ` -X "github.com/stampzilla/stampzilla-go/pkg/build.BuildTime=` + time.Now().Format(time.RFC3339) + `" -X github.com/stampzilla/stampzilla-go/pkg/build.Commit=` + os.Getenv("TRAVIS_COMMIT"),
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

func loadReleaseFromFile(file string) (*Release, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var release *Release

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&release)

	return release, err
}
