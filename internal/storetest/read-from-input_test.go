package storetest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadFromFile_PathResolution tests that paths are correctly resolved
// when reading test files, preventing path duplication issues.
//
// This test addresses the issue reported in:
// https://github.com/openfga/action-openfga-test/issues/32
//
// The bug occurred when the CLI's model test command would pass both:
// - fileName: the full path from glob (e.g., "my_project/infrastructure/fga/tests.fga.yaml")
// - basePath: path.Dir(fileName) (e.g., "my_project/infrastructure/fga")
//
// This caused ReadFromFile to join them, creating a duplicated path:
// "my_project/infrastructure/fga/my_project/infrastructure/fga/tests.fga.yaml"
//
// The fix is to pass an empty basePath when the fileName is already the complete
// path to the file (as returned by filepath.Glob). The file's directory is then
// used internally for resolving nested file references (model_file, tuple_file).
func TestReadFromFile_PathResolution(t *testing.T) {
	t.Parallel()

	// Create a nested directory structure similar to real-world usage
	// e.g., "my_project/infrastructure/fga/"
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "my_project", "infrastructure", "fga")
	err := os.MkdirAll(nestedDir, 0o755)
	require.NoError(t, err)

	// Create a simple model file
	modelContent := `model
  schema 1.1
type user

type document
  relations
    define viewer: [user]`
	modelFile := filepath.Join(nestedDir, "model.fga")
	err = os.WriteFile(modelFile, []byte(modelContent), 0o600)
	require.NoError(t, err)

	// Create a test file that references the model
	testContent := `name: Test Store
model_file: model.fga
tests:
  - name: test-1
    check:
      - user: user:anne
        object: document:doc1
        assertions:
          viewer: true`
	testFile := filepath.Join(nestedDir, "tests.fga.yaml")
	err = os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("with relative path from working directory", func(t *testing.T) {
		// Simulate the scenario where glob returns a relative path
		// like "my_project/infrastructure/fga/tests.fga.yaml"
		relPath := filepath.Join("my_project", "infrastructure", "fga", "tests.fga.yaml")

		// Change to temp directory to make the relative path valid
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// THIS IS THE BUG: Passing path.Dir(relPath) as basePath causes duplication
		// When fileName="my_project/infrastructure/fga/tests.fga.yaml"
		// and basePath="my_project/infrastructure/fga"
		// it tries to open "my_project/infrastructure/fga/my_project/infrastructure/fga/tests.fga.yaml"
		_, _, err = ReadFromFile(relPath, filepath.Dir(relPath))

		// This should fail with path duplication error
		assert.Error(t, err)
		if err != nil {
			// Verify the error message shows the duplicated path
			assert.Contains(t, err.Error(), "my_project/infrastructure/fga/my_project/infrastructure/fga")
		}
	})

	t.Run("with empty basePath - correct behavior", func(t *testing.T) {
		// The CORRECT way: pass empty basePath so the file path is used as-is
		relPath := filepath.Join("my_project", "infrastructure", "fga", "tests.fga.yaml")

		// Change to temp directory to make the relative path valid
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// With empty basePath, the file is opened at relPath directly
		// and nested references are resolved relative to the file's directory
		format, storeData, err := ReadFromFile(relPath, "")

		assert.NoError(t, err)
		assert.NotNil(t, storeData)
		assert.Equal(t, "Test Store", storeData.Name)
		assert.NotEmpty(t, storeData.Model)
		assert.Contains(t, storeData.Model, "type user")
		assert.NotNil(t, format)
	})

	t.Run("with absolute path", func(t *testing.T) {
		// When using absolute paths, basePath should be ignored
		absPath := testFile

		// Even with a non-empty basePath, absolute paths should work
		format, storeData, err := ReadFromFile(absPath, "/some/random/path")

		assert.NoError(t, err)
		assert.NotNil(t, storeData)
		assert.Equal(t, "Test Store", storeData.Name)
		assert.NotEmpty(t, storeData.Model)
		assert.NotNil(t, format)
	})

	t.Run("with nested file references", func(t *testing.T) {
		// Create a subdirectory with a model file
		subDir := filepath.Join(nestedDir, "models")
		err := os.MkdirAll(subDir, 0o755)
		require.NoError(t, err)

		subModelFile := filepath.Join(subDir, "submodel.fga")
		err = os.WriteFile(subModelFile, []byte(modelContent), 0o600)
		require.NoError(t, err)

		// Create a test file that references a model in a subdirectory
		testWithSubdirContent := `name: Test Store With Subdir
model_file: models/submodel.fga
tests:
  - name: test-1
    check:
      - user: user:bob
        object: document:doc2
        assertions:
          viewer: true`
		testFileWithSubdir := filepath.Join(nestedDir, "test-subdir.fga.yaml")
		err = os.WriteFile(testFileWithSubdir, []byte(testWithSubdirContent), 0o600)
		require.NoError(t, err)

		// Read with empty basePath - nested references should resolve correctly
		format, storeData, err := ReadFromFile(testFileWithSubdir, "")

		assert.NoError(t, err)
		assert.NotNil(t, storeData)
		assert.Equal(t, "Test Store With Subdir", storeData.Name)
		assert.NotEmpty(t, storeData.Model)
		assert.Contains(t, storeData.Model, "type user")
		assert.NotNil(t, format)
	})
}

// TestReadFromFile_ErrorMessages tests that error messages are clear and helpful.
func TestReadFromFile_ErrorMessages(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent file with path duplication shows resolved path", func(t *testing.T) {
		// This test verifies that when a file doesn't exist due to path duplication,
		// the error message shows the resolved (duplicated) path for debugging
		_, _, err := ReadFromFile("path/to/test.yaml", "path/to")

		require.Error(t, err)
		// The error should show both the original and resolved paths
		assert.Contains(t, strings.ToLower(err.Error()), "failed to read file")
		assert.Contains(t, err.Error(), "path/to/test.yaml")
		assert.Contains(t, err.Error(), "path/to/path/to/test.yaml")
	})
}
