package authorizationmodel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModFileAllowsParentPaths(t *testing.T) {
	t.Parallel()

	data := "schema: '1.2'\ncontents:\n  - ../core.fga\n  - wiki.fga\n"
	mod, err := parseModFile([]byte(data))
	require.NoError(t, err)
	require.Equal(t, "1.2", mod.Schema)
	require.Equal(t, []string{"../core.fga", "wiki.fga"}, mod.Contents)
}

func TestReadModelFromModFGAWithParentPaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// parent module
	corePath := filepath.Join(dir, "core.fga")
	require.NoError(t, os.WriteFile(corePath, []byte("module core\n  type user"), 0o600))

	// child directory with wiki module and modfile
	childDir := filepath.Join(dir, "child")
	require.NoError(t, os.Mkdir(childDir, 0o755))

	wikiPath := filepath.Join(childDir, "wiki.fga")
	require.NoError(t, os.WriteFile(wikiPath, []byte("module core\n  type wiki"), 0o600))

	modContents := "schema: '1.2'\ncontents:\n  - ../core.fga\n  - wiki.fga\n"
	modFile := filepath.Join(childDir, "fga.mod")
	require.NoError(t, os.WriteFile(modFile, []byte(modContents), 0o600))

	var model AuthzModel
	require.NoError(t, model.ReadModelFromModFGA(modFile))
	require.Len(t, model.GetTypeDefinitions(), 2)
}
