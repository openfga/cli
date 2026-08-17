//go:build unix

package model

import (
	"path/filepath"
	"sort"
	"syscall"
	"testing"
)

// Regression test for https://github.com/openfga/cli/issues/739
// A glob match that is a FIFO with no writer must never be handed to
// storetest.ReadFromFile, since os.ReadFile on such a FIFO blocks forever.
//
// This lives in a Unix-only file because syscall.Mkfifo does not exist on Windows, so a
// runtime GOOS skip would not prevent a compile failure there.
func TestResolveTestFilesSkipsNonRegularGlobMatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	regularPath := filepath.Join(dir, "ok.fga.yaml")
	if err := writeRegularFile(t, regularPath); err != nil {
		t.Fatalf("failed to write regular test file: %v", err)
	}

	fifoPath := filepath.Join(dir, "evil.fga.yaml")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	pattern := filepath.Join(dir, "*.fga.yaml")

	// This call must return promptly. If the FIFO were not filtered out, resolveTestFiles
	// itself doesn't read file contents, but a regression that moved the read here would hang
	// the test; test-unit's overall timeout is the real backstop for that case.
	fileNames, err := resolveTestFiles(pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fileNames) != 1 || fileNames[0] != regularPath {
		t.Fatalf("expected [%s], got %v", regularPath, fileNames)
	}
}

// Multiple regular files matched by the glob should all be returned, in addition to any
// non-regular files being dropped.
func TestResolveTestFilesMultipleRegularFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	first := filepath.Join(dir, "a.fga.yaml")
	second := filepath.Join(dir, "b.fga.yaml")

	for _, p := range []string{first, second} {
		if err := writeRegularFile(t, p); err != nil {
			t.Fatalf("failed to write regular test file: %v", err)
		}
	}

	if err := syscall.Mkfifo(filepath.Join(dir, "c.fga.yaml"), 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	pattern := filepath.Join(dir, "*.fga.yaml")

	fileNames, err := resolveTestFiles(pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort.Strings(fileNames)

	if len(fileNames) != 2 || fileNames[0] != first || fileNames[1] != second {
		t.Fatalf("expected [%s %s], got %v", first, second, fileNames)
	}
}

// If every glob match is non-regular, resolution must fail with a clear error rather than
// silently falling back to treating the glob pattern string itself as a literal path.
func TestResolveTestFilesAllMatchesNonRegular(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := syscall.Mkfifo(filepath.Join(dir, "evil.fga.yaml"), 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	pattern := filepath.Join(dir, "*.fga.yaml")

	fileNames, err := resolveTestFiles(pattern)
	if err == nil {
		t.Fatalf("expected an error, got fileNames=%v", fileNames)
	}
}
