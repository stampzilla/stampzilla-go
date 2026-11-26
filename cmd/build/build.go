package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var dest = "dist"

// Release is read from the RELEASE file and contains options for the build.
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
		build(path, release, true)
	}

	// The cli is not compressed since older versions of "stampzilla self-update"
	// and the install instructions in README.md download it by its plain name.
	fmt.Println("Building stampzilla cli")
	build("cmd/stampzilla", nil, false)
}

func build(path string, release *Release, compress bool) {
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

			args := getArgs(path, binName)
			fmt.Printf("building %s with args: %s\n", binName, args)
			cmd := exec.Command("go", args...) // #nosec
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED="+cgo)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(string(stdoutStderr))
				log.Fatal(err)
			}

			if compress {
				if err := gzipFile(filepath.Join(dest, binName)); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	fmt.Printf("done\n")
}

// gzipFile replaces file with a gzip compressed file named file.gz.
func gzipFile(file string) error {
	src, err := os.Open(file)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(file + ".gz")
	if err != nil {
		return err
	}
	defer dst.Close()

	gz, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := io.Copy(gz, src); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Remove(file)
}

func getArgs(path, binName string) []string {
	m := []string{
		"build",
		"-trimpath",
		"-ldflags",
		"-s -w -X github.com/stampzilla/stampzilla-go/v2/pkg/build.Version=" + version() + ` -X "github.com/stampzilla/stampzilla-go/v2/pkg/build.BuildTime=` + time.Now().Format(time.RFC3339) + `" -X github.com/stampzilla/stampzilla-go/v2/pkg/build.Commit=` + os.Getenv("GITHUB_SHA"),
		"-o",
		filepath.Join(dest, binName),
		"./" + path,
	}
	// filesToBuild, err := filepath.Glob(filepath.Join(path, "*.go"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// m = append(m, filesToBuild...)
	return m
}

// version returns the tag name when building from a tag in GitHub Actions.
func version() string {
	if os.Getenv("GITHUB_REF_TYPE") == "tag" {
		return os.Getenv("GITHUB_REF_NAME")
	}
	return ""
}

func loadReleaseFromFile(file string) (*Release, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var release *Release
	err = json.NewDecoder(f).Decode(&release)
	return release, err
}
