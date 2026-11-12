package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
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
