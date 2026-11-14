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
	// WalkDir does not descend into a symlinked root, so resolve the roots
	// first. Otherwise a symlinked config dir would give an empty backup.
	var err error
	if p.ConfigDir, err = resolveDir(p.ConfigDir); err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}
	if opts.IncludeBinaries {
		// a missing bin dir is reported per node below
		binDir, err := resolveDir(p.BinDir)
		switch {
		case err == nil:
			p.BinDir = binDir
		case !errors.Is(err, fs.ErrNotExist):
			return nil, fmt.Errorf("binary dir: %w", err)
		}
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

// resolveDir follows symlinks in dir and checks that the result is a directory.
func resolveDir(dir string) (string, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return resolved, nil
}

// ListNodes returns the installed stampzilla nodes: every node that has a
// binary in the binary dir or a config dir below ConfigDir/nodes.
func ListNodes(p Paths) ([]string, error) {
	seen := map[string]bool{}
	for _, src := range []struct {
		dir  string
		dirs bool
	}{
		{p.BinDir, false},
		{filepath.Join(p.ConfigDir, "nodes"), true},
	} {
		names, err := listNodeNames(src.dir, src.dirs)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		for _, n := range names {
			seen[n] = true
		}
	}
	nodes := make([]string, 0, len(seen))
	for n := range seen {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)
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
	// ConfigDir itself, once the whole bundle has been extracted. Children are
	// handed over before their parents and ConfigDir last. Optional.
	Chown func(path string) error
}

// staged is a binary written to a temp file next to its destination.
type staged struct{ tmp, dst string }

// Extract restores a bundle into p. ConfigDir is expected to be empty or
// missing and should be owned by root until Extract has returned.
//
// Directories are kept private (0700) while extracting so nobody can plant
// symlinks in them. Their archived modes are applied at the end together
// with the ownership handoff.
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

	var pending []staged
	defer func() {
		for _, s := range pending {
			os.Remove(s.tmp)
		}
	}()

	dirModes := map[string]fs.FileMode{}
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
			if hdr.Typeflag == tar.TypeDir && root == configPrefix {
				if err := os.MkdirAll(p.ConfigDir, 0o700); err != nil {
					return nil, err
				}
				dirModes[p.ConfigDir] = hdr.FileInfo().Mode().Perm()
				continue
			}
			if hdr.Typeflag == tar.TypeDir && root == binPrefix {
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
				if err := os.MkdirAll(dst, 0o700); err != nil {
					return nil, err
				}
				dirModes[dst] = mode
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
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

	// The tar reader stops at the end-of-archive marker without reading the
	// gzip trailer, so drain it to have the checksum verified.
	if _, err := io.Copy(io.Discard, gz); err != nil {
		return nil, fmt.Errorf("corrupt backup: %w", err)
	}

	if err := finalizeConfig(p.ConfigDir, dirModes, opts.Chown); err != nil {
		return nil, err
	}

	if err := commitBinaries(pending); err != nil {
		return nil, err
	}
	pending = nil
	return m, nil
}

// commitBinaries moves staged binaries into place. The previous binaries are
// kept as hard links until all of them have been replaced, so a failure
// halfway puts the old ones back instead of leaving a mix.
func commitBinaries(pending []staged) error {
	type done struct{ dst, old string }
	var committed []done
	rollback := func() {
		for i := len(committed) - 1; i >= 0; i-- {
			c := committed[i]
			if c.old == "" {
				os.Remove(c.dst)
				continue
			}
			os.Rename(c.old, c.dst)
		}
	}

	for _, s := range pending {
		old := ""
		fi, err := os.Lstat(s.dst)
		switch {
		case errors.Is(err, fs.ErrNotExist):
		case err != nil:
			rollback()
			return err
		case fi.IsDir():
			rollback()
			return fmt.Errorf("install %s: is a directory", s.dst)
		default:
			old, err = reserveName(filepath.Dir(s.dst), filepath.Base(s.dst)+".old")
			if err != nil {
				rollback()
				return err
			}
			if err := os.Link(s.dst, old); err != nil {
				rollback()
				return fmt.Errorf("keep previous %s: %w", s.dst, err)
			}
		}
		if err := os.Rename(s.tmp, s.dst); err != nil {
			if old != "" {
				os.Remove(old)
			}
			rollback()
			return fmt.Errorf("install %s: %w", s.dst, err)
		}
		committed = append(committed, done{dst: s.dst, old: old})
	}

	for _, c := range committed {
		if c.old != "" {
			os.Remove(c.old)
		}
	}
	return nil
}

// reserveName returns an unused path in dir for a hard link. The placeholder
// file is removed again, os.Link fails rather than follows if something else
// shows up at that name in the meantime.
func reserveName(dir, base string) (string, error) {
	f, err := os.CreateTemp(dir, "."+base+"-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	if err := os.Remove(name); err != nil {
		return "", err
	}
	return name, nil
}

// finalizeConfig applies directory modes and hands ownership of the config
// tree over. All paths are collected while the tree is still only writable by
// us, then processed children first so a directory is only given away once
// everything below it is done and its parent is still ours.
func finalizeConfig(root string, dirModes map[string]fs.FileMode, chown func(string) error) error {
	type entry struct {
		path string
		dir  bool
	}
	var entries []entry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		entries = append(entries, entry{path: path, dir: d.IsDir()})
		return nil
	})
	if err != nil {
		return fmt.Errorf("finalize config: %w", err)
	}
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if chown != nil {
			if err := chown(e.path); err != nil {
				return fmt.Errorf("chown config: %w", err)
			}
		}
		if e.dir {
			mode, ok := dirModes[e.path]
			if !ok {
				mode = 0o755
			}
			if err := os.Chmod(e.path, mode); err != nil {
				return fmt.Errorf("chmod config: %w", err)
			}
		}
	}
	return nil
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
