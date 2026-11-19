package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsInside(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "etc")
	for p, want := range map[string]bool{
		cfg:                                  true,
		filepath.Join(cfg, "x.tar.gz"):       true,
		filepath.Join(cfg, "a", "x.tar.gz"):  true,
		filepath.Join(dir, "etc2", "x"):      false,
		filepath.Join(dir, "x.tar.gz"):       false,
		filepath.Join(cfg, "..", "x.tar.gz"): false,
	} {
		got, err := isInside(p, cfg)
		require.NoError(t, err)
		assert.Equal(t, want, got, p)
	}
}
