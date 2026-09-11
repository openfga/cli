package mapping

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/mapper"
)

// noTTYConfirm simulates the non-TTY path: file exists, user must pass --force.
func noTTYConfirm(path string) (bool, error) {
	return false, fmt.Errorf("%s already exists — use --force to overwrite: %w", path, fs.ErrExist)
}

// yesConfirm simulates a user confirming the overwrite prompt.
func yesConfirm(_ string) (bool, error) { return true, nil }

// noConfirm simulates a user declining the overwrite prompt.
func noConfirm(_ string) (bool, error) { return false, nil }

// abortConfirm simulates the user pressing Ctrl+C.
func abortConfirm(_ string) (bool, error) { return false, huh.ErrUserAborted }

func TestInitMapping(t *testing.T) {
	t.Parallel()

	t.Run("creates mapping.yaml at given path", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")

		var out bytes.Buffer

		err := initMapping(path, false, false, &out)
		require.NoError(t, err)

		_, err = os.Stat(path)
		require.NoError(t, err, "file should exist")
		assert.Contains(t, out.String(), "Created")
		assert.Contains(t, out.String(), path)
	})

	t.Run("generated file passes mapper.Compile", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, initMapping(path, false, false, &bytes.Buffer{}))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = mapper.Compile(data)
		assert.NoError(t, err, "generated YAML must compile cleanly")
	})

	t.Run("starter template embedded tests pass", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, initMapping(path, false, false, &bytes.Buffer{}))

		err := runMappingTests(context.Background(), path, runMappingTestsOptions{}, &bytes.Buffer{}, &bytes.Buffer{})
		assert.NoError(t, err, "embedded tests in the starter template must all pass")
	})

	t.Run("--minimal generates a skeleton without tests block", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, initMapping(path, true, false, &bytes.Buffer{}))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "tests:")
		_, err = mapper.Compile(data)
		assert.NoError(t, err, "minimal YAML must compile cleanly")
	})

	t.Run("starter template contains tests block", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, initMapping(path, false, false, &bytes.Buffer{}))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "tests:")
	})

	t.Run("file exists without --force on non-tty returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, os.WriteFile(path, []byte("existing"), 0o600))

		err := initMappingWithConfirm(path, false, false, &bytes.Buffer{}, noTTYConfirm)
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrExist)
	})

	t.Run("file exists and user confirms overwrites with expected content", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))

		var out bytes.Buffer

		err := initMappingWithConfirm(path, false, false, &out, yesConfirm)
		require.NoError(t, err)

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, starterTemplate, string(data))
	})

	t.Run("file exists and user declines prints Aborted and exits clean", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, os.WriteFile(path, []byte("existing"), 0o600))

		var out bytes.Buffer

		err := initMappingWithConfirm(path, false, false, &out, noConfirm)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "Aborted.")

		// File must remain unchanged.
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "existing", string(data))
	})

	t.Run("user pressing Ctrl+C prints Aborted and exits clean", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, os.WriteFile(path, []byte("existing"), 0o600))

		var out bytes.Buffer

		err := initMappingWithConfirm(path, false, false, &out, abortConfirm)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "Aborted.")
	})

	t.Run("--force overwrites existing file with expected content", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "mapping.yaml")
		require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))
		require.NoError(t, initMapping(path, false, true, &bytes.Buffer{}))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, starterTemplate, string(data))
	})
}
