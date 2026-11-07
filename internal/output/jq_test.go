package output

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jqTestStore struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type jqTestResponse struct {
	Stores []jqTestStore `json:"stores"`
}

func TestSetJQFilter(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(".stores"))
	assert.NotNil(t, jqFilter)

	require.NoError(t, SetJQFilter(""))
	assert.Nil(t, jqFilter)

	err := SetJQFilter(".[")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid jq expression")
}

func TestApplyJQFilter(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	data := jqTestResponse{Stores: []jqTestStore{{ID: "01A", Name: "one"}, {ID: "01B", Name: "two"}}}

	tests := []struct {
		name       string
		expression string
		expected   []any
	}{
		{name: "field of a struct", expression: ".stores[0].id", expected: []any{"01A"}},
		{name: "one result per element", expression: ".stores[].name", expected: []any{"one", "two"}},
		{name: "object result", expression: ".stores[1]", expected: []any{map[string]any{"id": "01B", "name": "two"}}},
		{name: "numbers", expression: ".stores | length", expected: []any{2}},
		{name: "no results", expression: "empty", expected: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, SetJQFilter(test.expression))

			results, err := collectJQResults(data)
			require.NoError(t, err)
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestApplyJQFilterRuntimeError(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(".stores.id"))

	_, err := collectJQResults(jqTestResponse{Stores: []jqTestStore{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to apply jq expression")
}

func TestDisplayFilteredPrintsStringsRaw(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(".stores[] | .id, ."))

	var displayed []any

	err := displayFiltered(jqFilter, jqTestResponse{Stores: []jqTestStore{{ID: "01A", Name: "one"}}},
		func(data any) error {
			displayed = append(displayed, data)

			return nil
		})
	require.NoError(t, err)
	// The string result is printed directly, only the object goes through the display function.
	assert.Equal(t, []any{map[string]any{"id": "01A", "name": "one"}}, displayed)
}

func collectJQResults(data any) ([]any, error) {
	var results []any

	err := runJQFilter(jqFilter, data, func(result any) error {
		results = append(results, result)

		return nil
	})

	return results, err
}

func TestRunJQFilterEmitsResultsBeforeAnError(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(`.stores[].id, error("boom")`))

	results, err := collectJQResults(jqTestResponse{Stores: []jqTestStore{{ID: "01A"}}})
	require.Error(t, err)
	assert.Equal(t, []any{"01A"}, results)
}

func TestJQFilterReadsEnvironment(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })
	t.Setenv("FGA_JQ_TEST_STORE", "01B")

	require.NoError(t, SetJQFilter(`.stores[] | select(.id == env.FGA_JQ_TEST_STORE) | .name`))

	results, err := collectJQResults(jqTestResponse{Stores: []jqTestStore{{ID: "01A", Name: "one"}, {ID: "01B", Name: "two"}}})
	require.NoError(t, err)
	assert.Equal(t, []any{"two"}, results)
}

func TestCheckJQFilterApplied(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(""))
	require.NoError(t, CheckJQFilterApplied())

	require.NoError(t, SetJQFilter(".stores"))
	assert.True(t, JQFilterSet())
	require.ErrorIs(t, CheckJQFilterApplied(), ErrJQNotApplied)

	require.NoError(t, displayFiltered(jqFilter, jqTestResponse{}, func(any) error { return nil }))
	require.NoError(t, CheckJQFilterApplied())
}

func TestJQFilterKeepsLargeIntegers(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(".n, .n + 1"))

	results, err := collectJQResults(map[string]uint64{"n": 9007199254740993})
	require.NoError(t, err)
	require.Len(t, results, 2)

	encoded, err := json.Marshal(results)
	require.NoError(t, err)
	assert.JSONEq(t, "[9007199254740993, 9007199254740994]", string(encoded))
}

func TestJQFilterHalt(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(`.stores[].id, halt`))

	results, err := collectJQResults(jqTestResponse{Stores: []jqTestStore{{ID: "01A"}}})
	require.NoError(t, err)
	assert.Equal(t, []any{"01A"}, results)

	require.NoError(t, SetJQFilter(`null | halt_error(7)`))

	_, err = collectJQResults(jqTestResponse{})
	require.Error(t, err)
}

func TestUniPrinterSeparatesFilteredYAMLResults(t *testing.T) {
	t.Cleanup(func() { _ = SetJQFilter("") })

	require.NoError(t, SetJQFilter(".stores[]"))

	printed := captureStdout(t, func() {
		printer := UniPrinter{Printer: &yamlPrinter{}}
		require.NoError(t, printer.Display(jqTestResponse{Stores: []jqTestStore{{ID: "01A", Name: "one"}, {ID: "01B", Name: "two"}}}))
	})

	assert.Equal(t, "id: 01A\nname: one\n\n---\nid: 01B\nname: two\n\n", printed)
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	stdout := os.Stdout
	os.Stdout = writer

	t.Cleanup(func() { os.Stdout = stdout })

	run()

	require.NoError(t, writer.Close())

	os.Stdout = stdout

	printed, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(printed)
}
