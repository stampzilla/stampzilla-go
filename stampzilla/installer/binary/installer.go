package binary

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
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
	download(nodes)
}
func (t *Installer) Update(nodes ...string) {
	download(nodes, func(a github.ReleaseAsset) bool {
		_, err := os.Stat(getFilePath(a))
		return os.IsNotExist(err)
		// TODO: Compare to hash file in release
	})
}

// Download downloads a file and takes a callback. If callback returns true, skip download
func download(nodes []string, cb ...func(github.ReleaseAsset) bool) {
	releases := getReleases()

	if len(releases) < 1 {
		logrus.Error("No available releses found")
		return
	}

	checksums := getChecksums(releases[0].Assets)

outer:
	for _, v := range releases[0].Assets {
		// Skip all with wrong ARCH
		if !strings.HasSuffix(v.GetName(), runtime.GOARCH) {
			continue
		}

		// Skip nodes not requested from command line arguments
		inNodeList := false
		for _, n := range nodes {
			if strings.HasSuffix(v.GetName(), n+"-"+runtime.GOARCH) {
				inNodeList = true
			}
		}
		if !inNodeList && len(nodes) > 0 {
			continue
		}

		// Check if we should skip this one
		for _, c := range cb {
			if c(v) {
				continue outer
			}
		}

		//
		tmp, err := ioutil.TempFile("", "stampzilla")
		if err != nil {
			logrus.Fatal(err)
		}

		// Download the file
		err = fetch(v, tmp)
		if err != nil {
			logrus.Fatal(err)
		}

		// Validate checksum
		hasher := sha512.New()
		tmp.Seek(0, 0)
		if _, err := io.Copy(hasher, tmp); err != nil {
			logrus.Fatal(err)
		}
		checksum := fmt.Sprintf("%x", hasher.Sum(nil))
		tmp.Close()
		if checksums[v.GetName()] != checksum {
			logrus.WithFields(logrus.Fields{
				"expected": checksums[v.GetName()],
				"got":      checksum,
			}).Error("Wrong checksum for %s", v.GetName())
			os.Remove(tmp.Name())
		}

		// Move to installation dir
		err = os.Rename(tmp.Name(), getFilePath(v))
		if err != nil {
			logrus.Fatal(err)
		}

		//
		for k, n := range nodes {
			if strings.HasSuffix(v.GetName(), n+"-"+runtime.GOARCH) {
				nodes = append(nodes[:k], nodes[k+1:]...)
				break
			}
		}
	}

	for _, v := range nodes {
		logrus.Errorf("Failed to install %s", v)
	}
}

func getFilePath(ra github.ReleaseAsset) string {
	filename := strings.TrimSuffix(ra.GetName(), "-"+runtime.GOARCH)
	return filepath.Join(GetBinPath(), filename)
}

func fetch(v github.ReleaseAsset, file *os.File) error {
	logrus.Infof("Downloading %s from github.com", v.GetName())
	resp, err := http.Get(v.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	err = file.Chmod(0755)
	if err != nil {
		return err
	}

	return nil
}

func getChecksums(assets []github.ReleaseAsset) map[string]string {
	checksums := make(map[string]string)

	for _, v := range assets {
		if v.GetName() == "checksum" {
			resp, err := http.Get(v.GetBrowserDownloadURL())
			if err != nil {
				logrus.Fatal(err)
			}
			defer resp.Body.Close()

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.SplitN(scanner.Text(), " ", 2)
				if len(line) < 2 {
					continue
				}

				line[1] = strings.TrimSpace(line[1])
				checksums[filepath.Base(line[1])] = line[0]
			}
		}
	}

	return checksums
}

func GetBinPath() string {
	if runtime.GOOS == "linux" {
		return filepath.Join("/", "home", "stampzilla", "go", "bin")
	}

	logrus.Fatal("Unsupported os")
	return ""
}
