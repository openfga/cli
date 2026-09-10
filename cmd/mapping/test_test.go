package mapping

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOutput is a minimal struct for decoding JSON format output in tests.
type testOutput struct {
	Summary struct {
		Total  int `json:"total"`
		Passed int `json:"passed"`
		Failed int `json:"failed"`
	} `json:"summary"`
	Results []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"results"`
}

func TestRunMappingTests(t *testing.T) {
	t.Parallel()

	t.Run("all passing tests exits without error", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "json"}, &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result testOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.Equal(t, 2, result.Summary.Total)
		assert.Equal(t, 2, result.Summary.Passed)
		assert.Equal(t, 0, result.Summary.Failed)
		assert.Equal(t, "pass", result.Results[0].Status)
	})

	t.Run("result JSON is emitted even when tests fail", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/failing_tests.yaml",
			runMappingTestsOptions{format: "json"}, &out, &bytes.Buffer{})
		require.Error(t, err)

		var result testOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.Equal(t, 1, result.Summary.Total)
		assert.Equal(t, 0, result.Summary.Passed)
		assert.Equal(t, 1, result.Summary.Failed)
		assert.Equal(t, "fail", result.Results[0].Status)
	})

	t.Run("filter skips non-matching tests and reports filtered count", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "json", filter: "anne"}, &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result testOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		// Total counts all tests including filtered ones: 1 ran + 1 filtered = 2.
		assert.Equal(t, 2, result.Summary.Total)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "pass", result.Results[0].Status)
	})

	t.Run("filter matching no tests returns error", func(t *testing.T) {
		t.Parallel()

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{filter: "NOMATCH"}, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "matched no tests")
	})

	t.Run("results array is never null in JSON output", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		require.NoError(t, runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "json"}, &out, &bytes.Buffer{}))
		assert.Contains(t, out.String(), `"results": [`)
	})

	t.Run("fail-fast stops after first failure", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/failing_tests.yaml",
			runMappingTestsOptions{format: "json", failFast: true}, &out, &bytes.Buffer{})
		require.Error(t, err)

		var result testOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		assert.Equal(t, 1, result.Summary.Failed)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		t.Parallel()

		err := runMappingTests(context.Background(), "testdata/nonexistent.yaml",
			runMappingTestsOptions{}, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
	})

	t.Run("mapping with no tests returns error", func(t *testing.T) {
		t.Parallel()

		err := runMappingTests(context.Background(), "testdata/valid.yaml",
			runMappingTestsOptions{}, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no tests")
	})

	t.Run("text format writes human-readable output", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "text"}, &out, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "PASS")
		assert.Contains(t, out.String(), "2 passed")
	})

	t.Run("text format shows failure diff", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/failing_tests.yaml",
			runMappingTestsOptions{format: "text"}, &out, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, out.String(), "FAIL")
		assert.Contains(t, out.String(), "0 passed, 1 failed")
	})

	t.Run("junit format writes XML output", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "junit"}, &out, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "<?xml")
		assert.Contains(t, out.String(), "<testsuite")
	})

	t.Run("unknown format returns error", func(t *testing.T) {
		t.Parallel()

		err := runMappingTests(context.Background(), "testdata/with_tests.yaml",
			runMappingTestsOptions{format: "xml"}, &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown format")
	})

	t.Run("compile error returns errMappingInvalid", func(t *testing.T) {
		t.Parallel()

		err := runMappingTests(context.Background(), "testdata/invalid.yaml",
			runMappingTestsOptions{}, &bytes.Buffer{}, &bytes.Buffer{})
		require.ErrorIs(t, err, errMappingInvalid)
	})
}
