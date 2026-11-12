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
			if _, err := os.Stat(filepath.Join(p.BinDir, n)); err != nil {
				return nil, fmt.Errorf("binary for %s: %w", n, err)
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
		if !strings.HasPrefix(e.Name(), "stampzilla-") || strings.HasSuffix(e.Name(), ".new") {
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
	return m, nil
}

// RestoreOptions controls how a bundle is extracted.
type RestoreOptions struct {
	// SkipBinaries skips extraction of bundled binaries.
	SkipBinaries bool
	// Chown is called for every extracted config file and directory. Optional.
	Chown func(path string) error
}

// Extract restores a bundle into p. ConfigDir is expected to be empty or missing.
func Extract(r io.Reader, p Paths, opts RestoreOptions) (*Manifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a stampzilla backup: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

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

		var dstRoot string
		var chown func(string) error
		switch root {
		case configPrefix:
			dstRoot = p.ConfigDir
			chown = opts.Chown
		case binPrefix:
			if opts.SkipBinaries {
				continue
			}
			dstRoot = p.BinDir
		default:
			return nil, fmt.Errorf("unexpected entry in backup: %q", hdr.Name)
		}

		dst, err := safeJoin(dstRoot, rel)
		if err != nil {
			return nil, err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return nil, err
			}
			if err := os.Chmod(dst, hdr.FileInfo().Mode().Perm()); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return nil, err
			}
			if err := writeFile(dst, tr, hdr.FileInfo().Mode().Perm()); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported entry type %q for %s", hdr.Typeflag, hdr.Name)
		}

		if chown != nil {
			if err := chown(dst); err != nil {
				return nil, err
			}
		}
		_ = os.Chtimes(dst, hdr.ModTime, hdr.ModTime)
	}

	if m == nil {
		return nil, fmt.Errorf("not a stampzilla backup: empty archive")
	}
	return m, nil
}

// writeFile writes to a temporary file and renames it into place so running
// binaries can be replaced without "text file busy" errors.
func writeFile(dst string, r io.Reader, mode fs.FileMode) error {
	tmp := dst + ".new"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Chmod(tmp, mode); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
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
