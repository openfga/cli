package mapping

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptMappingFile(t *testing.T) {
	t.Parallel()

	// An explicit path argument is returned verbatim without prompting. The
	// no-argument branches (TTY picker, non-interactive usage error) call
	// os.Exit and are exercised via the built binary, not here.
	t.Run("returns an explicit path argument unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "mapping.yaml", promptMappingFile([]string{"mapping.yaml"}, io.Discard))
	})
}
