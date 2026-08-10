//go:build unix

package safefile

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOpenedFIFOIsRejectedWithoutBlocking simulates the race where a path
// passes the metadata check as a regular file but is a FIFO by the time it is
// opened: the nonblocking open must return immediately rather than wait for a
// writer, and the descriptor check must reject the file before any read.
func TestOpenedFIFOIsRejectedWithoutBlocking(t *testing.T) {
	t.Parallel()

	base := t.TempDir()

	if err := syscall.Mkfifo(filepath.Join(base, "pipe"), 0o600); err != nil {
		t.Skipf("unable to create FIFO: %v", err)
	}

	root, err := os.OpenRoot(base)
	require.NoError(t, err)

	defer root.Close()

	// A blocking open would hang here forever: no writer ever opens the FIFO.
	file, err := root.OpenFile("pipe", os.O_RDONLY|openNonblock, 0)
	require.NoError(t, err)

	defer file.Close()

	_, err = readOpened("pipe", file)
	require.ErrorIs(t, err, ErrNotRegularFile)
}
