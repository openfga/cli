package storetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadRefContainment verifies that references escaping the base directory
// are rejected by default but permitted when external files are explicitly
// allowed.
func TestReadRefContainment(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	base := filepath.Join(root, "base")
	require.NoError(t, os.Mkdir(base, 0o750))

	// A regular file inside the base directory.
	inside := writeTempFile(t, base, "model.fga", "model")

	// A regular file one level above the base directory (the traversal target).
	writeTempFile(t, root, "secret.txt", "SENSITIVE")

	t.Run("reference inside base resolves and reads", func(t *testing.T) {
		t.Parallel()

		path, contents, err := readRef(base, filepath.Base(inside), false)
		require.NoError(t, err)
		assert.Equal(t, inside, path)
		assert.Equal(t, "model", string(contents))
	})

	t.Run("traversal is blocked by default", func(t *testing.T) {
		t.Parallel()

		_, _, err := readRef(base, "../secret.txt", false)
		require.Error(t, err)
	})

	t.Run("absolute path is blocked by default", func(t *testing.T) {
		t.Parallel()

		_, _, err := readRef(base, filepath.Join(root, "secret.txt"), false)
		require.Error(t, err)
	})

	t.Run("traversal is allowed when external files are permitted", func(t *testing.T) {
		t.Parallel()

		path, contents, err := readRef(base, "../secret.txt", true)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "secret.txt"), path)
		assert.Equal(t, "SENSITIVE", string(contents))
	})
}

// TestReadRefSymlinkTraversalIsBlocked covers the case where a lexical path and
// a root-relative lookup disagree.
//
// With a symlink "link" -> "sub/dir", os.Root resolves "link/../target.json" to
// "sub/target.json" while filepath.Join collapses it to "target.json". If the
// containment check ran against the former but the read used the latter, a
// "target.json" symlink pointing outside the base directory would be followed
// and its contents returned. The read therefore has to go through the same
// os.Root handle the check used.
func TestReadRefSymlinkTraversalIsBlocked(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	base := filepath.Join(root, "base")
	require.NoError(t, os.MkdirAll(filepath.Join(base, "sub", "dir"), 0o750))

	// The file the attacker wants, outside the base directory.
	writeTempFile(t, root, "outside-secret.json", "EXFILTRATED")

	// base/link -> sub/dir, so "link/.." resolves to base/sub through os.Root but
	// collapses to base lexically.
	require.NoError(t, os.Symlink(filepath.Join("sub", "dir"), filepath.Join(base, "link")))

	// The lexical target: base/target.json, itself a symlink out of the tree.
	require.NoError(t, os.Symlink(
		filepath.Join("..", "outside-secret.json"),
		filepath.Join(base, "target.json"),
	))

	// Something harmless exists at the root-relative target, so a check that
	// resolves through the root succeeds and only the read location differs.
	writeTempFile(t, filepath.Join(base, "sub"), "target.json", "harmless")

	_, contents, err := readRef(base, "link/../target.json", false)

	// Either the reference is rejected outright, or it resolves through the root
	// to the harmless in-tree file. What must never happen is reading the file
	// outside the base directory.
	if err == nil {
		assert.NotContains(t, string(contents), "EXFILTRATED",
			"read escaped the base directory via symlink")
		assert.Equal(t, "harmless", string(contents))
	}
}

// TestReadRefRejectsDirectory verifies a reference to a directory is rejected
// rather than read.
func TestReadRefRejectsDirectory(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(base, "adir"), 0o750))

	_, _, err := readRef(base, "adir", false)
	require.Error(t, err)

	_, _, err = readRef(base, "adir", true)
	require.Error(t, err)
}
