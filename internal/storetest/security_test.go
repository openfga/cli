package storetest

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveFileContainment verifies that references escaping the base
// directory are rejected by default but permitted when external files are
// explicitly allowed.
func TestResolveFileContainment(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	base := filepath.Join(root, "base")
	require.NoError(t, os.Mkdir(base, 0o750))

	// A regular file inside the base directory.
	inside := writeTempFile(t, base, "model.fga", "model")

	// A regular file one level above the base directory (the traversal target).
	writeTempFile(t, root, "secret.txt", "SENSITIVE")

	t.Run("reference inside base resolves", func(t *testing.T) {
		t.Parallel()

		got, err := resolveFile(base, filepath.Base(inside), false)
		require.NoError(t, err)
		assert.Equal(t, inside, got)
	})

	t.Run("traversal is blocked by default", func(t *testing.T) {
		t.Parallel()

		_, err := resolveFile(base, "../secret.txt", false)
		require.Error(t, err)
	})

	t.Run("absolute path is blocked by default", func(t *testing.T) {
		t.Parallel()

		_, err := resolveFile(base, filepath.Join(root, "secret.txt"), false)
		require.Error(t, err)
	})

	t.Run("traversal is allowed when external files are permitted", func(t *testing.T) {
		t.Parallel()

		got, err := resolveFile(base, "../secret.txt", true)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "secret.txt"), got)
	})
}

// TestResolveFileRejectsNonRegular verifies that FIFOs (and other special
// files) are rejected before any read, guarding against the DoS (a FIFO read
// blocks forever; an infinite device like /dev/zero exhausts memory).
func TestResolveFileRejectsNonRegular(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mkfifo is not available on Windows")
	}

	t.Parallel()

	base := t.TempDir()
	fifo := filepath.Join(base, "pipe")

	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	// Contained mode: os.Root.Stat rejects the FIFO as non-regular.
	_, err := resolveFile(base, filepath.Base(fifo), false)
	require.Error(t, err)

	// External-files mode: safefile.CheckRegular still rejects the FIFO.
	_, err = resolveFile(base, filepath.Base(fifo), true)
	require.Error(t, err)
}
