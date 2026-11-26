package binary

import (
	"bytes"
	"compress/gzip"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func asset(name string) github.ReleaseAsset {
	return github.ReleaseAsset{Name: &name}
}

func TestAssetNames(t *testing.T) {
	suffix := "-" + runtime.GOOS + "-" + runtime.GOARCH

	plain := asset("stampzilla-server" + suffix)
	gz := asset("stampzilla-server" + suffix + ".gz")

	assert.False(t, isGzip(plain))
	assert.True(t, isGzip(gz))
	assert.Equal(t, "stampzilla-server"+suffix, binaryName(plain))
	assert.Equal(t, "stampzilla-server"+suffix, binaryName(gz))
	assert.Equal(t, filepath.Join(GetBinPath(), "stampzilla-server"), getFilePath(plain))
	assert.Equal(t, filepath.Join(GetBinPath(), "stampzilla-server"), getFilePath(gz))
}

func TestWriteBinary(t *testing.T) {
	content := []byte("\x7fELF binary content")

	var plain bytes.Buffer
	assert.NoError(t, writeBinary(&plain, bytes.NewReader(content), false))
	assert.Equal(t, content, plain.Bytes())

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	_, err := gz.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, gz.Close())

	var out bytes.Buffer
	assert.NoError(t, writeBinary(&out, &compressed, true))
	assert.Equal(t, content, out.Bytes())

	// Not gzip data
	assert.Error(t, writeBinary(&out, bytes.NewReader(content), true))
}

func TestSplitVersion(t *testing.T) {
	tests := []struct{ in, name, version string }{
		{"tibber", "tibber", ""},
		{"stampzilla-tibber", "stampzilla-tibber", ""},
		{"tibber@v2.9.0", "tibber", "v2.9.0"},
		{"tibber@2.9.0", "tibber", "v2.9.0"},
		{"tibber@", "tibber", ""},
	}
	for _, tt := range tests {
		name, version := SplitVersion(tt.in)
		assert.Equal(t, tt.name, name, tt.in)
		assert.Equal(t, tt.version, version, tt.in)
	}
}

func TestGroupByVersion(t *testing.T) {
	versions, groups := groupByVersion([]string{"server", "tibber@v2.9.0", "deconz", "wled@v2.9.0", "knx@v2.8.2"})
	assert.Equal(t, []string{"", "v2.9.0", "v2.8.2"}, versions)
	assert.Equal(t, []string{"server", "deconz"}, groups[""])
	assert.Equal(t, []string{"tibber", "wled"}, groups["v2.9.0"])
	assert.Equal(t, []string{"knx"}, groups["v2.8.2"])
}
