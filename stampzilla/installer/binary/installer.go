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

	"github.com/google/go-github/github"
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
	if len(nodes) == 0 {
		return fmt.Errorf("Please specifiy which nodes you like to install. (ex 'stampzilla install server example')")
	}
	return download(nodes, func(a github.ReleaseAsset) bool {
		_, err := os.Stat(getFilePath(a))
		if err == nil {
			logrus.Infof("%s already exists. use stampzilla install -u to update", a.GetName())
			return true
		}
		return false
	})
}
func (t *Installer) Update(nodes ...string) error {
	return download(nodes, func(a github.ReleaseAsset) bool {
		_, err := os.Stat(getFilePath(a))
		if os.IsNotExist(err) {
			logrus.Infof("%s does not exist. use stampzilla install to install", a.GetName())
			return true
		}
		return false
	})
}

// Download downloads a file and takes a callback. If callback returns true, skip download
func download(nodes []string, cb ...func(github.ReleaseAsset) bool) error {
	releases := getReleases()

	if len(releases) < 1 {
		return fmt.Errorf("No available releses found")
	}

	checksums := getChecksums(releases[0].Assets)

	expectedToInstall := make([]string, len(nodes))
	copy(expectedToInstall, nodes)

outer:
	for _, v := range releases[0].Assets {
		// Skip all with wrong ARCH
		if !strings.HasSuffix(v.GetName(), runtime.GOOS+"-"+runtime.GOARCH) {
			continue
		}

		// Skip nodes not requested from command line arguments
		inNodeList := false
		for _, n := range nodes {
			if strings.HasSuffix(v.GetName(), n+"-"+runtime.GOOS+"-"+runtime.GOARCH) {
				inNodeList = true
				break
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
			return err
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// Download the file
		err = fetch(v, *releases[0].TagName, tmp)
		if err != nil {
			return err
		}

		// Validate checksum
		hasher := sha512.New()
		tmp.Seek(0, 0)
		if _, err := io.Copy(hasher, tmp); err != nil {
			return err
		}
		checksum := fmt.Sprintf("%x", hasher.Sum(nil))
		if checksums[v.GetName()] != checksum {
			logrus.WithFields(logrus.Fields{
				"expected": checksums[v.GetName()],
				"got":      checksum,
			}).Errorf("Wrong checksum for %s", v.GetName())
			return fmt.Errorf("checksum validation failed")
		}

		// Move to installation dir
		dst, err := os.OpenFile(getFilePath(v), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		tmp.Seek(0, 0)
		_, err = io.Copy(dst, tmp)
		if err != nil {
			return err
		}

		//

		for k, n := range expectedToInstall {
			if strings.HasSuffix(v.GetName(), n+"-"+runtime.GOOS+"-"+runtime.GOARCH) {
				expectedToInstall = append(expectedToInstall[:k], expectedToInstall[k+1:]...)
				break
			}
		}
	}

	if len(expectedToInstall) > 0 {
		return fmt.Errorf("Failed to install %s", strings.Join(expectedToInstall, ","))
	}
	return nil
}

func getFilePath(ra github.ReleaseAsset) string {
	filename := strings.TrimSuffix(ra.GetName(), "-"+runtime.GOOS+"-"+runtime.GOARCH)
	return filepath.Join(GetBinPath(), filename)
}

func fetch(v github.ReleaseAsset, version string, file *os.File) error {
	logrus.Infof("Downloading %s@%s from github.com", v.GetName(), version)
	resp, err := http.Get(v.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return file.Chmod(0755)
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
