package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, p, content string, mode os.FileMode) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), mode))
}

func setupSource(t *testing.T) Paths {
	t.Helper()
	dir := t.TempDir()
	p := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	writeTestFile(t, filepath.Join(p.ConfigDir, "nodes.conf"), `{"daemons":[]}`, 0o644)
	writeTestFile(t, filepath.Join(p.ConfigDir, "nodes", "stampzilla-server", "config.json"), `{"port":"8080"}`, 0o644)
	writeTestFile(t, filepath.Join(p.ConfigDir, "nodes", "stampzilla-server", "certificates", "ca.key"), "secret", 0o600)
	writeTestFile(t, filepath.Join(p.ConfigDir, "nodes", "stampzilla-knx", "config.json"), `{}`, 0o644)
	writeTestFile(t, filepath.Join(p.BinDir, "stampzilla-server"), "server-binary", 0o755)
	writeTestFile(t, filepath.Join(p.BinDir, "stampzilla-knx"), "knx-binary", 0o755)
	writeTestFile(t, filepath.Join(p.BinDir, "stampzilla-knx.new"), "partial", 0o755)
	writeTestFile(t, filepath.Join(p.BinDir, "other"), "not ours", 0o755)
	return p
}

func TestBackupRestore(t *testing.T) {
	src := setupSource(t)

	buf := &bytes.Buffer{}
	m, err := Create(buf, src, Options{IncludeBinaries: true, Version: "v1.2.3"})
	require.NoError(t, err)
	assert.Equal(t, []string{"stampzilla-knx", "stampzilla-server"}, m.Nodes)

	read, err := ReadManifest(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", read.Version)
	assert.True(t, read.Binaries)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	chowned := 0
	m2, err := Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{Chown: func(string) error { chowned++; return nil }})
	require.NoError(t, err)
	assert.Equal(t, m.Nodes, m2.Nodes)

	data, err := os.ReadFile(filepath.Join(dst.ConfigDir, "nodes", "stampzilla-server", "config.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"port":"8080"}`, string(data))

	fi, err := os.Stat(filepath.Join(dst.ConfigDir, "nodes", "stampzilla-server", "certificates", "ca.key"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())

	fi, err = os.Stat(filepath.Join(dst.BinDir, "stampzilla-server"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), fi.Mode().Perm())

	assert.NoFileExists(t, filepath.Join(dst.BinDir, "other"))
	assert.NoFileExists(t, filepath.Join(dst.BinDir, "stampzilla-knx.new"))
	assert.Greater(t, chowned, 5)
}

func TestBackupWithoutBinaries(t *testing.T) {
	src := setupSource(t)
	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{})
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dst.ConfigDir, "nodes.conf"))
	assert.NoDirExists(t, dst.BinDir)
}

func TestListNodesFallsBackToConfigDirs(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.RemoveAll(src.BinDir))
	nodes, err := ListNodes(src)
	require.NoError(t, err)
	assert.Equal(t, []string{"stampzilla-knx", "stampzilla-server"}, nodes)
}

func TestListNodesMergesBinariesAndConfigDirs(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.Remove(filepath.Join(src.BinDir, "stampzilla-knx")))
	require.NoError(t, os.RemoveAll(filepath.Join(src.ConfigDir, "nodes", "stampzilla-server")))
	writeTestFile(t, filepath.Join(src.BinDir, "stampzilla-tibber"), "tibber-binary", 0o755)
	nodes, err := ListNodes(src)
	require.NoError(t, err)
	assert.Equal(t, []string{"stampzilla-knx", "stampzilla-server", "stampzilla-tibber"}, nodes)
}

func TestBackupBinariesDetectsPartiallyMissingBinaries(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.Remove(filepath.Join(src.BinDir, "stampzilla-knx")))
	_, err := Create(&bytes.Buffer{}, src, Options{IncludeBinaries: true})
	assert.ErrorContains(t, err, "binary for stampzilla-knx")
}

func TestExtractRejectsBadGzipChecksum(t *testing.T) {
	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	writeTestFile(t, filepath.Join(dst.BinDir, "stampzilla-server"), "old", 0o755)

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "config/", typ: tar.TypeDir},
		{name: "bin/stampzilla-server", content: "new"},
	}, false)
	data[len(data)-8] ^= 0xff // CRC32 in the gzip trailer
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{})
	assert.ErrorContains(t, err, "checksum")

	got, err := os.ReadFile(filepath.Join(dst.BinDir, "stampzilla-server"))
	require.NoError(t, err)
	assert.Equal(t, "old", string(got))
}

func TestExtractRestoresDirModes(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.Chmod(src.ConfigDir, 0o700))
	require.NoError(t, os.Chmod(filepath.Join(src.ConfigDir, "nodes", "stampzilla-knx"), 0o750))
	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{})
	require.NoError(t, err)

	for p, want := range map[string]os.FileMode{
		dst.ConfigDir: 0o700,
		filepath.Join(dst.ConfigDir, "nodes", "stampzilla-knx"):    0o750,
		filepath.Join(dst.ConfigDir, "nodes", "stampzilla-server"): 0o755,
	} {
		fi, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, want, fi.Mode().Perm(), p)
	}
}

func TestExtractKeepsDirsPrivateUntilDone(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.Chmod(filepath.Join(src.ConfigDir, "nodes"), 0o777))
	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	nodesDir := filepath.Join(dst.ConfigDir, "nodes")
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{Chown: func(p string) error {
		if p == nodesDir {
			// the mode is applied right after the dir has been handed over
			fi, err := os.Stat(p)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o700), fi.Mode().Perm())
		}
		return nil
	}})
	require.NoError(t, err)
	fi, err := os.Stat(nodesDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), fi.Mode().Perm())
}

func TestExtractRejectsPathTraversal(t *testing.T) {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, writeManifest(tw, &Manifest{FormatVersion: FormatVersion}))
	content := []byte("evil")
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "config/../../evil", Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "a", "etc"), BinDir: filepath.Join(dir, "a", "bin")}
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{})
	assert.ErrorContains(t, err, "invalid path")
	assert.NoFileExists(t, filepath.Join(dir, "evil"))
}

func TestReadManifestRejectsOtherArchives(t *testing.T) {
	_, err := ReadManifest(bytes.NewReader([]byte("hello")))
	assert.Error(t, err)
}

func TestBackupBinariesRequiresBinaries(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.RemoveAll(src.BinDir))
	_, err := Create(&bytes.Buffer{}, src, Options{IncludeBinaries: true})
	assert.ErrorContains(t, err, "binary for stampzilla-knx")
}

func TestBackupWarnsAboutSymlinks(t *testing.T) {
	src := setupSource(t)
	require.NoError(t, os.Symlink("/nonexistent", filepath.Join(src.ConfigDir, "link")))
	warnings := []string{}
	_, err := Create(&bytes.Buffer{}, src, Options{Warn: func(s string) { warnings = append(warnings, s) }})
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "link")
}

type testEntry struct {
	name    string
	content string
	typ     byte
}

// buildArchive builds a bundle by hand. If truncate is set the gzip stream is cut off.
func buildArchive(t *testing.T, m *Manifest, entries []testEntry, truncate bool) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, writeManifest(tw, m))
	for _, e := range entries {
		typ := e.typ
		if typ == 0 {
			typ = tar.TypeReg
		}
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.content)), Typeflag: typ}))
		_, err := tw.Write([]byte(e.content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	b := buf.Bytes()
	if truncate {
		b = b[:len(b)-30]
	}
	return b
}

func TestManifestRejectsInvalidNodeNames(t *testing.T) {
	for _, name := range []string{"stampzilla-/../../tmp/pwn", "other", "stampzilla-a b", "stampzilla-x.new", ""} {
		data := buildArchive(t, &Manifest{FormatVersion: FormatVersion, Nodes: []string{name}}, nil, false)
		_, err := ReadManifest(bytes.NewReader(data))
		assert.ErrorContains(t, err, "invalid node name", name)
	}
}

func TestExtractRejectsNestedBinaries(t *testing.T) {
	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{{name: "bin/link/stampzilla-x", content: "x"}}, false)
	dir := t.TempDir()
	_, err := Extract(bytes.NewReader(data), Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}, RestoreOptions{})
	assert.ErrorContains(t, err, "unexpected binary entry")
}

func TestCorruptBundleLeavesBinariesUntouched(t *testing.T) {
	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	writeTestFile(t, filepath.Join(dst.BinDir, "stampzilla-server"), "old", 0o755)

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "bin/stampzilla-server", content: "new"},
		{name: "config/big", content: string(bytes.Repeat([]byte("x"), 100000))},
	}, true)
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{})
	require.Error(t, err)

	got, err := os.ReadFile(filepath.Join(dst.BinDir, "stampzilla-server"))
	require.NoError(t, err)
	assert.Equal(t, "old", string(got))
	entries, err := os.ReadDir(dst.BinDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "staged temp files should be removed")
}

func TestBackupBinariesRejectsSymlinkedBinary(t *testing.T) {
	src := setupSource(t)
	knx := filepath.Join(src.BinDir, "stampzilla-knx")
	require.NoError(t, os.Remove(knx))
	require.NoError(t, os.Symlink(filepath.Join(src.BinDir, "other"), knx))
	_, err := Create(&bytes.Buffer{}, src, Options{IncludeBinaries: true})
	assert.ErrorContains(t, err, "not a regular file")
}

func TestExtractChownsWholeTreeAfterwards(t *testing.T) {
	src := setupSource(t)
	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	chowned := map[string]bool{}
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{Chown: func(p string) error {
		chowned[p] = true
		return nil
	}})
	require.NoError(t, err)
	assert.True(t, chowned[dst.ConfigDir])
	assert.True(t, chowned[filepath.Join(dst.ConfigDir, "nodes", "stampzilla-server", "certificates", "ca.key")])
}

func TestExtractChownsChildrenBeforeParents(t *testing.T) {
	src := setupSource(t)
	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	var order []string
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{Chown: func(p string) error {
		order = append(order, p)
		return nil
	}})
	require.NoError(t, err)
	require.NotEmpty(t, order)
	assert.Equal(t, dst.ConfigDir, order[len(order)-1])
	done := map[string]bool{}
	for _, p := range order {
		for parent := filepath.Dir(p); strings.HasPrefix(parent, dst.ConfigDir); parent = filepath.Dir(parent) {
			assert.False(t, done[parent], "%s chowned before its child %s", parent, p)
		}
		done[p] = true
	}
}

func TestBackupFollowsSymlinkedConfigDir(t *testing.T) {
	src := setupSource(t)
	link := filepath.Join(t.TempDir(), "etc-link")
	require.NoError(t, os.Symlink(src.ConfigDir, link))
	src.ConfigDir = link

	buf := &bytes.Buffer{}
	_, err := Create(buf, src, Options{})
	require.NoError(t, err)

	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	_, err = Extract(bytes.NewReader(buf.Bytes()), dst, RestoreOptions{})
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(dst.ConfigDir, "nodes", "stampzilla-server", "config.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"port":"8080"}`, string(got))
}

func TestBackupRejectsConfigDirThatIsAFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "etc")
	writeTestFile(t, f, "", 0o644)
	_, err := Create(&bytes.Buffer{}, Paths{ConfigDir: f, BinDir: t.TempDir()}, Options{})
	assert.ErrorContains(t, err, "not a directory")
}

func TestFailedBinaryCommitRestoresPreviousBinaries(t *testing.T) {
	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	writeTestFile(t, filepath.Join(dst.BinDir, "stampzilla-a"), "old-a", 0o755)
	// a directory where the second binary should go makes its commit fail
	require.NoError(t, os.MkdirAll(filepath.Join(dst.BinDir, "stampzilla-b"), 0o755))

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "config/", typ: tar.TypeDir},
		{name: "bin/stampzilla-a", content: "new-a"},
		{name: "bin/stampzilla-b", content: "new-b"},
		{name: "bin/stampzilla-c", content: "new-c"},
	}, false)
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{})
	require.ErrorContains(t, err, "is a directory")

	got, err := os.ReadFile(filepath.Join(dst.BinDir, "stampzilla-a"))
	require.NoError(t, err)
	assert.Equal(t, "old-a", string(got))
	entries, err := os.ReadDir(dst.BinDir)
	require.NoError(t, err)
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	assert.ElementsMatch(t, []string{"stampzilla-a", "stampzilla-b"}, names, "no temp files or new binaries should be left")
}

func TestBinaryCommitReplacesExistingBinaries(t *testing.T) {
	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	writeTestFile(t, filepath.Join(dst.BinDir, "stampzilla-a"), "old-a", 0o755)

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "config/", typ: tar.TypeDir},
		{name: "bin/stampzilla-a", content: "new-a"},
		{name: "bin/stampzilla-b", content: "new-b"},
	}, false)
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{})
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(dst.BinDir, "stampzilla-a"))
	require.NoError(t, err)
	assert.Equal(t, "new-a", string(got))
	entries, err := os.ReadDir(dst.BinDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "previous binaries kept during commit should be removed")
}

func TestBinaryCommitDoesNotFollowSymlinkedBinary(t *testing.T) {
	dir := t.TempDir()
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	outside := filepath.Join(dir, "outside")
	writeTestFile(t, outside, "important", 0o644)
	require.NoError(t, os.MkdirAll(dst.BinDir, 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(dst.BinDir, "stampzilla-a")))

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "config/", typ: tar.TypeDir},
		{name: "bin/stampzilla-a", content: "new-a"},
	}, false)
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{})
	require.NoError(t, err)

	got, err := os.ReadFile(outside)
	require.NoError(t, err)
	assert.Equal(t, "important", string(got), "file behind the symlink must not be touched")
	fi, err := os.Lstat(filepath.Join(dst.BinDir, "stampzilla-a"))
	require.NoError(t, err)
	assert.True(t, fi.Mode().IsRegular(), "symlink should be replaced by the binary")
}

func TestExtractRefusesBinDirRejectedByCheck(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "elsewhere")
	require.NoError(t, os.MkdirAll(target, 0o755))
	dst := Paths{ConfigDir: filepath.Join(dir, "etc"), BinDir: filepath.Join(dir, "bin")}
	require.NoError(t, os.Symlink(target, dst.BinDir))

	data := buildArchive(t, &Manifest{FormatVersion: FormatVersion}, []testEntry{
		{name: "config/", typ: tar.TypeDir},
		{name: "bin/stampzilla-a", content: "new-a"},
	}, false)
	var checked os.FileInfo
	_, err := Extract(bytes.NewReader(data), dst, RestoreOptions{CheckBinDir: func(fi os.FileInfo) error {
		checked = fi
		return assert.AnError
	}})
	require.ErrorIs(t, err, assert.AnError)
	require.NotNil(t, checked)
	assert.True(t, checked.IsDir(), "check should get the directory the symlink points to")
	entries, err := os.ReadDir(target)
	require.NoError(t, err)
	assert.Empty(t, entries, "nothing should be written to a refused bin dir")
}
