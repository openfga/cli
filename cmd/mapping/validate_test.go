package mapping

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMapping(t *testing.T) {
	t.Parallel()

	t.Run("text mode prints human summary on success", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := validateMapping("testdata/valid.yaml", "text", "", false, &out, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "valid")
		assert.Contains(t, out.String(), "1 rules")
	})

	t.Run("json mode reports valid and rule count", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := validateMapping("testdata/valid.yaml", "json", "", false, &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result validateResult
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.True(t, result.Valid)
		assert.Equal(t, 1, result.RuleCount)
	})

	t.Run("invalid mapping text mode writes diagnostics to stderr", func(t *testing.T) {
		t.Parallel()

		var errOut bytes.Buffer

		err := validateMapping("testdata/invalid.yaml", "text", "", false, &bytes.Buffer{}, &errOut)
		require.Error(t, err)
		assert.Contains(t, errOut.String(), "unclosed")
	})

	t.Run("invalid mapping json mode writes structured errors to stdout", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := validateMapping("testdata/invalid.yaml", "json", "", false, &out, &bytes.Buffer{})
		require.Error(t, err)

		var result validateResult
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("unknown format returns error", func(t *testing.T) {
		t.Parallel()

		err := validateMapping("testdata/valid.yaml", "xml", "", false, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown format")
	})

	t.Run("missing file returns error", func(t *testing.T) {
		t.Parallel()

		err := validateMapping("testdata/nonexistent.yaml", "text", "", false, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "reading")
	})

	t.Run("model-file validates consistent mapping", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := validateMapping("testdata/valid.yaml", "text", "testdata/model.fga", false, &out, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "valid")
	})

	t.Run("model-file text mode reports inconsistency to stderr", func(t *testing.T) {
		t.Parallel()

		var errOut bytes.Buffer

		// valid.yaml references relation "member" on type "org" — consistent with model.fga.
		// Use a mapping that references a non-existent relation.
		err := validateMapping("testdata/bad_relation.yaml", "text", "testdata/model.fga", false, &bytes.Buffer{}, &errOut)
		require.ErrorIs(t, err, errModelInconsistent)
		assert.Contains(t, errOut.String(), "relation")
	})

	t.Run("model-file json mode reports model_errors in output", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := validateMapping("testdata/bad_relation.yaml", "json", "testdata/model.fga", false, &out, &bytes.Buffer{})
		require.ErrorIs(t, err, errModelInconsistent)

		var result validateResult
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.ModelErrors)
	})

	t.Run("missing model-file returns error", func(t *testing.T) {
		t.Parallel()

		err := validateMapping("testdata/valid.yaml", "text", "testdata/nonexistent.fga", false, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "loading model")
	})
}
