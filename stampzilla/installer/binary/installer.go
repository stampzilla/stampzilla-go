package binary

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
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
	download()
}
func (t *Installer) Update(nodes ...string) {
	download(func(a github.ReleaseAsset, f string) bool {
		_, err := os.Stat(f)
		return os.IsNotExist(err)
		// TODO: Compare to hash file in release
	})
}

// Download downloads a file and takes a callback. If callback returns true, skip download
func download(cb ...func(github.ReleaseAsset, string) bool) {
	releases := getReleases()

	if len(releases) < 1 {
		logrus.Error("No available releses found")
		return
	}

outer:
	for _, v := range releases[0].Assets {
		if !strings.HasSuffix(v.GetName(), runtime.GOARCH) {
			continue
		}

		filename := strings.TrimSuffix(v.GetName(), "-"+runtime.GOARCH)
		filename = filepath.Join(GetBinPath(), filename)

		// Check if we should skip this one
		for _, c := range cb {
			if c(v, filename) {
				continue outer
			}
		}

		logrus.Infof("Downloading %s from github.com", v.GetName())
		resp, err := http.Get(v.GetBrowserDownloadURL())
		if err != nil {
			logrus.Fatal(err)
		}
		defer resp.Body.Close()

		file, err := os.Create(filename)
		//file, err := os.OpenFile(filepath.Join(GetBinPath(), filename), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			logrus.Fatal(err)
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			logrus.Fatal(err)
		}

		err = os.Chmod(filename, 0755)
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func GetBinPath() string {
	if runtime.GOOS == "linux" {
		return filepath.Join("/", "home", "stampzilla", "go", "bin")
	}

	logrus.Fatal("Unsupported os")
	return ""
}
