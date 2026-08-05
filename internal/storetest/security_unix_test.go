//go:build unix

package storetest

import (
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestReadRefRejectsNonRegular verifies that FIFOs (and other special files) are
// rejected before any read. A FIFO with no writer would block the read forever;
// an endless device such as /dev/zero would grow the read buffer until the
// process is OOM-killed.
//
// This lives in a Unix-only file because syscall.Mkfifo does not exist on
// Windows, so a runtime GOOS skip would not prevent a compile failure there.
func TestReadRefRejectsNonRegular(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	fifo := filepath.Join(base, "pipe")

	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	// Contained mode: rejected via the os.Root lookup.
	_, _, err := readRef(base, filepath.Base(fifo), false)
	require.Error(t, err)

	// External-files mode: still rejected, containment is not what catches this.
	_, _, err = readRef(base, filepath.Base(fifo), true)
	require.Error(t, err)
}
