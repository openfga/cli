package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeRegularFile writes a minimal regular test file at path. Shared by both this file and
// test_unix_test.go.
func writeRegularFile(t *testing.T, path string) error {
	t.Helper()

	if err := os.WriteFile(path, []byte("name: ok\n"), 0o600); err != nil {
		return fmt.Errorf("failed to write test file %s: %w", path, err)
	}

	return nil
}

// A pattern with no glob matches at all is treated as a literal path. Since the pattern here
// contains a wildcard and no such literal file exists, resolution should fail with a
// not-found error rather than silently returning an empty result.
func TestResolveTestFilesNoMatchesTreatedAsLiteralPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pattern := filepath.Join(dir, "*.fga.yaml")

	fileNames, err := resolveTestFiles(pattern)
	if err == nil {
		t.Fatalf("expected an error, got fileNames=%v", fileNames)
	}
}

// An existing literal path (no glob metacharacters, so filepath.Glob returns it as a single
// match) is returned as-is, without the regular-file check rejecting it - this preserves
// existing behavior for explicitly named paths.
func TestResolveTestFilesLiteralExistingPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ok.fga.yaml")

	if err := writeRegularFile(t, path); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fileNames, err := resolveTestFiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fileNames) != 1 || fileNames[0] != path {
		t.Fatalf("expected [%s], got %v", path, fileNames)
	}
}
