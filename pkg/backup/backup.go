// Package backup creates and restores stampzilla backup bundles.
//
// A bundle is a gzipped tar archive with the following layout:
//
//	manifest.json   metadata about the backup (always the first entry)
//	config/...      the contents of the config dir (/etc/stampzilla)
//	bin/...         node binaries (optional)
package backup

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	FormatVersion = 1

	manifestName = "manifest.json"
	configPrefix = "config"
	binPrefix    = "bin"
)

// Paths points out where stampzilla keeps its files.
type Paths struct {
	ConfigDir string // usually /etc/stampzilla
	BinDir    string // usually /home/stampzilla/go/bin
}

// Manifest describes a backup bundle.
type Manifest struct {
	FormatVersion int       `json:"formatVersion"`
	CreatedAt     time.Time `json:"createdAt"`
	Hostname      string    `json:"hostname"`
	Version       string    `json:"version"`
	OS            string    `json:"os"`
	Arch          string    `json:"arch"`
	Nodes         []string  `json:"nodes"`
	Binaries      bool      `json:"binaries"`
}

// Options controls what goes into a backup.
type Options struct {
	IncludeBinaries bool
	Version         string
	// Warn is called for files that are skipped. Optional.
	Warn func(msg string)
}

// Create writes a backup bundle to w.
func Create(w io.Writer, p Paths, opts Options) (*Manifest, error) {
	if _, err := os.Stat(p.ConfigDir); err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}

	nodes, err := ListNodes(p)
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	m := &Manifest{
		FormatVersion: FormatVersion,
		CreatedAt:     time.Now().UTC(),
		Hostname:      hostname,
		Version:       opts.Version,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		Nodes:         nodes,
	}

	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)

	if opts.IncludeBinaries {
		for _, n := range nodes {
			fi, err := os.Lstat(filepath.Join(p.BinDir, n))
			if err != nil {
				return nil, fmt.Errorf("binary for %s: %w", n, err)
			}
			if !fi.Mode().IsRegular() {
				return nil, fmt.Errorf("binary for %s is not a regular file (%s)", n, fi.Mode().Type())
			}
		}
		m.Binaries = len(nodes) > 0
	}

	if err := writeManifest(tw, m); err != nil {
		return nil, err
	}

	if err := addDir(tw, p.ConfigDir, configPrefix, nil, opts.Warn); err != nil {
		return nil, fmt.Errorf("add config: %w", err)
	}

	if opts.IncludeBinaries {
		wanted := map[string]bool{}
		for _, n := range nodes {
			wanted[n] = true
		}
		filter := func(rel string, d fs.DirEntry) bool {
			if d.IsDir() {
				return rel == "."
			}
			return wanted[rel]
		}
		if err := addDir(tw, p.BinDir, binPrefix, filter, opts.Warn); err != nil {
			return nil, fmt.Errorf("add binaries: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return m, nil
}

// ListNodes returns the installed stampzilla nodes. It looks in the binary dir
// first and falls back to the node config dirs if no binaries are found.
func ListNodes(p Paths) ([]string, error) {
	nodes, err := listNodeNames(p.BinDir, false)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if len(nodes) > 0 {
		return nodes, nil
	}
	nodes, err = listNodeNames(filepath.Join(p.ConfigDir, "nodes"), true)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return nodes, nil
}

func listNodeNames(dir string, dirs bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, e := range entries {
		if e.IsDir() != dirs {
			continue
		}
		if !validNodeName.MatchString(e.Name()) { // also skips installer leftovers like stampzilla-x.new
			continue
		}
		ret = append(ret, e.Name())
	}
	sort.Strings(ret)
	return ret, nil
}

func writeManifest(tw *tar.Writer, m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name:    manifestName,
		Mode:    0o644,
		Size:    int64(len(data)),
		ModTime: m.CreatedAt,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = tw.Write(data)
	return err
}

// addDir adds all directories and regular files below root to the archive
// under prefix. If filter is set, entries it returns false for are skipped.
func addDir(tw *tar.Writer, root, prefix string, filter func(rel string, d fs.DirEntry) bool, warn func(string)) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if filter != nil && !filter(rel, d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			// sockets, symlinks, devices etc are not supported
			if warn != nil {
				warn(fmt.Sprintf("skipping %s: not a regular file (%s)", p, info.Mode().Type()))
			}
			return nil
		}

		name := path.Join(prefix, rel)
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = name
		// ownership is set on restore, the uid/gid will differ between machines
		hdr.Uid, hdr.Gid, hdr.Uname, hdr.Gname = 0, 0, "", ""
		if info.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		// copy exactly hdr.Size bytes so a file growing during backup can't corrupt the archive
		if _, err := io.CopyN(tw, f, hdr.Size); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		return nil
	})
}

// ReadManifest reads the manifest from a bundle without extracting it.
func ReadManifest(r io.Reader) (*Manifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a stampzilla backup: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	hdr, err := tr.Next()
	if err != nil {
		return nil, fmt.Errorf("not a stampzilla backup: %w", err)
	}
	if hdr.Name != manifestName {
		return nil, fmt.Errorf("not a stampzilla backup: first entry is %q, expected %s", hdr.Name, manifestName)
	}
	return decodeManifest(tr)
}

func decodeManifest(r io.Reader) (*Manifest, error) {
	m := &Manifest{}
	if err := json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	if m.FormatVersion < 1 || m.FormatVersion > FormatVersion {
		return nil, fmt.Errorf("unsupported backup format version %d", m.FormatVersion)
	}
	for _, n := range m.Nodes {
		if !validNodeName.MatchString(n) {
			return nil, fmt.Errorf("invalid node name in manifest: %q", n)
		}
	}
	return m, nil
}

var validNodeName = regexp.MustCompile(`^stampzilla-[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// RestoreOptions controls how a bundle is extracted.
type RestoreOptions struct {
	// SkipBinaries skips extraction of bundled binaries.
	SkipBinaries bool
	// Chown is called for every extracted config file and directory, including
	// ConfigDir itself, once the whole bundle has been extracted. Optional.
	Chown func(path string) error
}

// Extract restores a bundle into p. ConfigDir is expected to be empty or
// missing and should be owned by root until Extract has returned.
//
// Binaries are staged next to their destination and only moved into place
// once the whole bundle has been read successfully, so a corrupt bundle
// leaves the existing binaries untouched.
func Extract(r io.Reader, p Paths, opts RestoreOptions) (*Manifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a stampzilla backup: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	type staged struct{ tmp, dst string }
	var pending []staged
	defer func() {
		for _, s := range pending {
			os.Remove(s.tmp)
		}
	}()

	var m *Manifest
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		if m == nil {
			if hdr.Name != manifestName {
				return nil, fmt.Errorf("not a stampzilla backup: first entry is %q, expected %s", hdr.Name, manifestName)
			}
			m, err = decodeManifest(tr)
			if err != nil {
				return nil, err
			}
			continue
		}

		root, rel, ok := strings.Cut(strings.TrimSuffix(hdr.Name, "/"), "/")
		if !ok {
			if hdr.Typeflag == tar.TypeDir && (root == configPrefix || root == binPrefix) {
				continue // top level dir
			}
			return nil, fmt.Errorf("unexpected entry in backup: %q", hdr.Name)
		}

		switch root {
		case configPrefix:
			dst, err := safeJoin(p.ConfigDir, rel)
			if err != nil {
				return nil, err
			}
			mode := hdr.FileInfo().Mode().Perm()
			switch hdr.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(dst, 0o755); err != nil {
					return nil, err
				}
				if err := os.Chmod(dst, mode); err != nil {
					return nil, err
				}
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
					return nil, err
				}
				tmp, err := writeTemp(filepath.Dir(dst), filepath.Base(dst), tr, mode)
				if err != nil {
					return nil, err
				}
				if err := os.Rename(tmp, dst); err != nil {
					os.Remove(tmp)
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unsupported entry type %q for %s", hdr.Typeflag, hdr.Name)
			}
			_ = os.Chtimes(dst, hdr.ModTime, hdr.ModTime)

		case binPrefix:
			if opts.SkipBinaries {
				continue
			}
			// binaries are flat regular files named after the node
			if hdr.Typeflag != tar.TypeReg || !validNodeName.MatchString(rel) {
				return nil, fmt.Errorf("unexpected binary entry in backup: %q", hdr.Name)
			}
			if len(pending) == 0 {
				if err := os.MkdirAll(p.BinDir, 0o755); err != nil {
					return nil, err
				}
				fi, err := os.Lstat(p.BinDir)
				if err != nil {
					return nil, err
				}
				if !fi.IsDir() {
					return nil, fmt.Errorf("%s is not a directory", p.BinDir)
				}
			}
			tmp, err := writeTemp(p.BinDir, rel, tr, hdr.FileInfo().Mode().Perm())
			if err != nil {
				return nil, err
			}
			pending = append(pending, staged{tmp: tmp, dst: filepath.Join(p.BinDir, rel)})

		default:
			return nil, fmt.Errorf("unexpected entry in backup: %q", hdr.Name)
		}
	}

	if m == nil {
		return nil, fmt.Errorf("not a stampzilla backup: empty archive")
	}

	// Ownership is handed over last so the config tree can't be tampered
	// with (e.g. swapping a directory for a symlink) while root writes to it.
	if opts.Chown != nil {
		err := filepath.WalkDir(p.ConfigDir, func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return opts.Chown(path)
		})
		if err != nil {
			return nil, fmt.Errorf("chown config: %w", err)
		}
	}

	for len(pending) > 0 {
		s := pending[0]
		if err := os.Rename(s.tmp, s.dst); err != nil {
			return nil, err
		}
		pending = pending[1:]
	}
	return m, nil
}

// writeTemp writes r to a new, exclusively created temporary file in dir and
// returns its path. Renaming it into place lets running binaries be replaced
// without "text file busy" errors.
func writeTemp(dir, base string, r io.Reader, mode fs.FileMode) (string, error) {
	f, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	if err := f.Chmod(mode); err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return tmp, nil
}

func safeJoin(root, rel string) (string, error) {
	if rel == "" || path.IsAbs(rel) || strings.Contains(rel, "\\") {
		return "", fmt.Errorf("invalid path in backup: %q", rel)
	}
	clean := path.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("invalid path in backup: %q", rel)
	}
	return filepath.Join(root, filepath.FromSlash(clean)), nil
}
