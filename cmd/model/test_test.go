package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandTestFilePatterns_SingleFile(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{"../../example/model.fga.yaml"}, []string{})
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Contains(t, files[0], "example/model.fga.yaml")
}

func TestExpandTestFilePatterns_MultipleFilesWithFlag(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns(
		[]string{"../../example/model.fga.yaml", "../../example/store_abac.fga.yaml"},
		[]string{},
	)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/store_abac.fga.yaml"))
}

func TestExpandTestFilePatterns_GlobPattern(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{"../../example/*.fga.yaml"}, []string{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2, "should match at least model.fga.yaml and store_abac.fga.yaml")
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/store_abac.fga.yaml"))
}

func TestExpandTestFilePatterns_ShellExpandedFiles(t *testing.T) {
	t.Parallel()

	// Simulate what happens when the shell expands: --tests file1 file2
	// The first file goes to the flag, the second becomes a positional arg
	files, err := expandTestFilePatterns(
		[]string{"../../example/model.fga.yaml"},
		[]string{"../../example/store_abac.fga.yaml"},
	)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/store_abac.fga.yaml"))
}

func TestExpandTestFilePatterns_MixedGlobAndLiteral(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns(
		[]string{"../../example/model.fga.yaml", "../../example/*.fga.yaml"},
		[]string{},
	)
	require.NoError(t, err)
	// Should include model.fga.yaml and store_abac.fga.yaml
	assert.GreaterOrEqual(t, len(files), 2, "should have at least 2 files")
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/store_abac.fga.yaml"))
}

func TestExpandTestFilePatterns_SubdirectoryGlob(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{"../../example/subdir/*.fga.yaml"}, []string{})
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.True(t, anyContains(files, "example/subdir/simple.fga.yaml"))
}

func TestExpandTestFilePatterns_NonExistentFile(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{"../../example/nonexistent.fga.yaml"}, []string{})
	require.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "does not exist", "error should mention file doesn't exist")
}

func TestExpandTestFilePatterns_NoMatchingFiles(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{"../../example/*.nonexistent"}, []string{})
	require.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "does not exist", "error should mention no files found")
}

func TestExpandTestFilePatterns_NoFilesSpecified(t *testing.T) {
	t.Parallel()

	files, err := expandTestFilePatterns([]string{}, []string{})
	require.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "no test files specified")
}

func TestExpandTestFilePatterns_ShellExpandedThreeFiles(t *testing.T) {
	t.Parallel()

	// Simulate shell expanding multiple globs that result in 3 files
	files, err := expandTestFilePatterns(
		[]string{"../../example/model.fga.yaml"},
		[]string{"../../example/store_abac.fga.yaml", "../../example/subdir/simple.fga.yaml"},
	)
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/store_abac.fga.yaml"))
	assert.True(t, anyContains(files, "example/subdir/simple.fga.yaml"))
}

func TestExpandTestFilePatterns_GlobPatternNotExpandedByShell(t *testing.T) {
	t.Parallel()

	// When user quotes the pattern, shell doesn't expand it
	// So the CLI receives the glob pattern itself
	files, err := expandTestFilePatterns([]string{"../../example/*.fga.yaml"}, []string{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2, "should expand glob to at least 2 files")
}

func TestExpandTestFilePatterns_CombineGlobAndShellExpanded(t *testing.T) {
	t.Parallel()

	// Mixed scenario: one file and one glob pattern as positional arg
	files, err := expandTestFilePatterns(
		[]string{"../../example/model.fga.yaml"},
		[]string{"../../example/subdir/*.fga.yaml"},
	)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.True(t, anyContains(files, "example/model.fga.yaml"))
	assert.True(t, anyContains(files, "example/subdir/simple.fga.yaml"))
}

func TestExpandTestFilePatterns_InvalidGlobPattern(t *testing.T) {
	t.Parallel()

	// Test with an invalid glob pattern (contains invalid characters for glob)
	files, err := expandTestFilePatterns([]string{"../../example/[.fga.yaml"}, []string{})
	require.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "invalid tests pattern")
}

// anyContains checks if any string in the slice contains the given substring.
func anyContains(slice []string, substr string) bool {
	for _, s := range slice {
		if strings.Contains(s, substr) {
			return true
		}
	}

	return false
}
