package model

import (
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"testing"
)

// Regression test for https://github.com/openfga/cli/issues/739
// A glob match that is a FIFO with no writer must never be handed to
// storetest.ReadFromFile, since os.ReadFile on such a FIFO blocks forever.
func TestResolveTestFilesSkipsNonRegularGlobMatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	regularPath := filepath.Join(dir, "ok.fga.yaml")
	if err := os.WriteFile(regularPath, []byte("name: ok\n"), 0o600); err != nil {
		t.Fatalf("failed to write regular test file: %v", err)
	}

	fifoPath := filepath.Join(dir, "evil.fga.yaml")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Fatalf("failed to create fifo: %v", err)
	}

	pattern := filepath.Join(dir, "*.fga.yaml")

	// This call must return promptly. If the FIFO were not filtered out, resolveTestFiles
	// itself doesn't read file contents, but a regression that moved the read here would hang
	// the test; test-unit's overall timeout is the real backstop for that case.
	fileNames, err := resolveTestFiles(pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fileNames) != 1 {
		t.Fatalf("expected exactly 1 regular file, got %d: %v", len(fileNames), fileNames)
	}

	if fileNames[0] != regularPath {
		t.Fatalf("expected %s, got %s", regularPath, fileNames[0])
	}
}

// A pattern with no glob matches at all is treated as a literal path by the caller, not by
// resolveTestFiles itself - so an empty result here is expected and not an error.
func TestResolveTestFilesNoMatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pattern := filepath.Join(dir, "*.fga.yaml")

	fileNames, err := resolveTestFiles(pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fileNames) != 0 {
		t.Fatalf("expected no matches, got %v", fileNames)
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
		if err := os.WriteFile(p, []byte("name: ok\n"), 0o600); err != nil {
			t.Fatalf("failed to write regular test file: %v", err)
		}
	}

	if err := syscall.Mkfifo(filepath.Join(dir, "c.fga.yaml"), 0o600); err != nil {
		t.Fatalf("failed to create fifo: %v", err)
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
