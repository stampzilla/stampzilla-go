package binary

import (
	"bufio"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"io"
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
		return fmt.Errorf("Please specify which nodes you like to install. (ex 'stampzilla install server example')")
	}
	return downloadVersions(nodes, func(a github.ReleaseAsset) bool {
		_, err := os.Stat(getFilePath(a))
		if err == nil {
			logrus.Infof("%s already exists. use stampzilla update to update", getFilePath(a))
			return true
		}
		return false
	})
}

func (t *Installer) Update(nodes ...string) error {
	return downloadVersions(nodes, func(a github.ReleaseAsset) bool {
		_, err := os.Stat(getFilePath(a))
		if os.IsNotExist(err) {
			logrus.Infof("%s does not exist. use stampzilla install to install", getFilePath(a))
			return true
		}
		return false
	})
}

// SplitVersion splits a node argument like "tibber@v2.9.0" into name and
// release version. The version is empty if none was given, which means latest.
func SplitVersion(node string) (name, version string) {
	name, version, _ = strings.Cut(node, "@")
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return name, version
}

// groupByVersion groups node names by requested release version, keeping
// the order in which versions first appear.
func groupByVersion(nodes []string) ([]string, map[string][]string) {
	versions := []string{}
	groups := make(map[string][]string)
	for _, n := range nodes {
		name, version := SplitVersion(n)
		if _, ok := groups[version]; !ok {
			versions = append(versions, version)
		}
		groups[version] = append(groups[version], name)
	}
	return versions, groups
}

// downloadVersions downloads nodes, each from its requested release (ex tibber@v2.9.0) or the latest one.
func downloadVersions(nodes []string, cb ...func(github.ReleaseAsset) bool) error {
	if len(nodes) == 0 {
		return download("", nil, cb...)
	}
	versions, groups := groupByVersion(nodes)
	for _, version := range versions {
		if err := download(version, groups[version], cb...); err != nil {
			return err
		}
	}
	return nil
}

// Download downloads a file and takes a callback. If callback returns true, skip download.
// An empty version means the latest release.
func download(version string, nodes []string, cb ...func(github.ReleaseAsset) bool) error {
	release, err := getRelease(version)
	if err != nil {
		return err
	}

	checksums, err := getChecksums(release.Assets)
	if err != nil {
		return err
	}

	expectedToInstall := make([]string, len(nodes))
	copy(expectedToInstall, nodes)

	logrus.Infof("looking for binaries with suffix: %s", runtime.GOOS+"-"+runtime.GOARCH)
outer:
	for _, v := range release.Assets {
		logrus.Debugf("checking asset: %s %s", v.GetName(), v.GetBrowserDownloadURL())
		name := binaryName(v)
		// Skip all with wrong ARCH
		if !strings.HasSuffix(name, runtime.GOOS+"-"+runtime.GOARCH) {
			continue
		}

		// Skip nodes not requested from command line arguments
		inNodeList := false
		for _, n := range nodes {
			if strings.HasSuffix(name, n+"-"+runtime.GOOS+"-"+runtime.GOARCH) {
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

		if err := installAsset(v, release.GetTagName(), checksums); err != nil {
			return err
		}

		for k, n := range expectedToInstall {
			if strings.HasSuffix(name, n+"-"+runtime.GOOS+"-"+runtime.GOARCH) {
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

// installAsset downloads, validates and installs one release asset.
func installAsset(v github.ReleaseAsset, version string, checksums map[string]string) error {
	tmp, err := os.CreateTemp("", "stampzilla")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	// Download the file
	err = fetch(v, version, tmp)
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

	// Move to temp file in installation dir
	dst, err := os.OpenFile(getFilePath(v)+".new", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	tmp.Seek(0, 0)
	err = writeBinary(dst, tmp, isGzip(v))
	// Close can report delayed write errors, never install a file that failed to close.
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(getFilePath(v) + ".new")
		return fmt.Errorf("error writing %s: %w", getFilePath(v), err)
	}

	// Do atomic rename to avoid file busy errors
	err = os.Rename(getFilePath(v)+".new", getFilePath(v))
	if err != nil {
		os.Remove(getFilePath(v) + ".new")
		return err
	}

	return nil
}

// isGzip reports whether the release asset is a gzip compressed binary.
func isGzip(ra github.ReleaseAsset) bool {
	return strings.HasSuffix(ra.GetName(), ".gz")
}

// binaryName returns the asset name without compression suffix, ex stampzilla-server-linux-amd64.
func binaryName(ra github.ReleaseAsset) string {
	return strings.TrimSuffix(ra.GetName(), ".gz")
}

// writeBinary copies src to dst, decompressing it first if compressed is true.
func writeBinary(dst io.Writer, src io.Reader, compressed bool) error {
	if compressed {
		gz, err := gzip.NewReader(src)
		if err != nil {
			return err
		}
		defer gz.Close()
		src = gz
	}
	_, err := io.Copy(dst, src)
	return err
}

func getFilePath(ra github.ReleaseAsset) string {
	filename := strings.TrimSuffix(binaryName(ra), "-"+runtime.GOOS+"-"+runtime.GOARCH)
	return filepath.Join(GetBinPath(), filename)
}

func fetch(v github.ReleaseAsset, version string, file *os.File) error {
	logrus.Infof("Downloading %s@%s from github.com", v.GetName(), version)
	resp, err := http.Get(v.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading %s: %s", v.GetName(), resp.Status)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return file.Chmod(0755)
}

func getChecksums(assets []github.ReleaseAsset) (map[string]string, error) {
	checksums := make(map[string]string)

	for _, v := range assets {
		if v.GetName() == "checksum" {
			resp, err := http.Get(v.GetBrowserDownloadURL())
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("error downloading checksum: %s", resp.Status)
			}

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.SplitN(scanner.Text(), " ", 2)
				if len(line) < 2 {
					continue
				}

				line[1] = strings.TrimSpace(line[1])
				checksums[filepath.Base(line[1])] = line[0]
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
		}
	}

	return checksums, nil
}

func GetBinPath() string {
	if runtime.GOOS == "linux" {
		return filepath.Join("/", "home", "stampzilla", "go", "bin")
	}

	logrus.Fatal("Unsupported os")
	return ""
}
