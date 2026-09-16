package mapping

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testOpts(format string, writesOnly bool) runMappingOptions {
	return runMappingOptions{format: format, writesOnly: writesOnly}
}

func TestRunMapping(t *testing.T) {
	t.Parallel()

	t.Run("emits write tuple as JSONL with op field", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false), strings.NewReader(`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var raw map[string]any
		require.NoError(t, json.NewDecoder(&out).Decode(&raw))
		assert.Equal(t, "user:anne", raw["user"])
		assert.Equal(t, "member", raw["relation"])
		assert.Equal(t, "org:acme", raw["object"])
		assert.Equal(t, "write", raw["op"])
		assert.NotContains(t, raw, "action")
	})

	t.Run("default JSONL emits both writes and deletes with op", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/mixed_actions.yaml", testOpts("jsonl", false), strings.NewReader(`{"id":"anne","org":"acme"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)

		ops := make(map[string]string)

		for _, line := range lines {
			var op opTupleOutput
			require.NoError(t, json.Unmarshal([]byte(line), &op))
			ops[op.Relation] = op.Op
		}

		assert.Equal(t, "write", ops["member"])
		assert.Equal(t, "delete", ops["viewer"])
	})

	t.Run("writes-only JSONL emits ClientTupleKey shape (no op or action)", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", true), strings.NewReader(`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var tuple writeTupleOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&tuple))
		assert.Equal(t, "user:anne", tuple.User)
		assert.Equal(t, "member", tuple.Relation)
		assert.Equal(t, "org:acme", tuple.Object)
		assert.Nil(t, tuple.Condition)

		assert.NotContains(t, out.String(), `"op"`)
		assert.NotContains(t, out.String(), `"action"`)
	})

	t.Run("writes-only filters out delete tuples", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/mixed_actions.yaml", testOpts("jsonl", true), strings.NewReader(`{"id":"anne","org":"acme"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], `"member"`)
		assert.NotContains(t, lines[0], `"viewer"`)
	})

	t.Run("json full format returns batch object splitting writes and deletes", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/mixed_actions.yaml", testOpts("json", false), strings.NewReader(`{"id":"anne","org":"acme"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.Writes, 1)
		require.Len(t, result.Deletes, 1)
		assert.Equal(t, "member", result.Writes[0].Relation)
		assert.Equal(t, "viewer", result.Deletes[0].Relation)
		assert.NotNil(t, result.UnresolvedFilters)
	})

	t.Run("json full format lists unresolved filters and does not warn", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/with_filter.yaml", testOpts("json", false), strings.NewReader(`{"id":"anne","org":"acme"}`), &out, &errOut)
		require.NoError(t, err)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.UnresolvedFilters, 1)
		assert.Equal(t, "user:anne", result.UnresolvedFilters[0].User)
		assert.Equal(t, "org:acme", result.UnresolvedFilters[0].Object)
		assert.Empty(t, errOut.String())
	})

	t.Run("json writes-only emits flat array consumable by fga tuple write", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("json", true), strings.NewReader(`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var writes []writeTupleOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&writes))
		require.Len(t, writes, 1)
		assert.Equal(t, "user:anne", writes[0].User)
	})

	t.Run("json writes-only marshals as [] not null when empty", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		// delete_only.yaml produces no writes — writes-only yields an empty set.
		// Assert on raw bytes: decoding cannot distinguish null from [], but
		// fga tuple write --file requires a JSON array.
		err := runMapping(context.Background(), "testdata/delete_only.yaml", testOpts("json", true), strings.NewReader(`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		assert.Equal(t, "[]", strings.TrimSpace(out.String()))
	})

	t.Run("json batch marshals empty writes/deletes/filters as [] not null", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		// delete_only.yaml yields a delete and no writes; the batch must still
		// render writes as an empty array and unresolved_filters as [].
		err := runMapping(context.Background(), "testdata/delete_only.yaml", testOpts("json", false), strings.NewReader(`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		raw := out.String()
		assert.Contains(t, raw, `"writes":[]`)
		assert.Contains(t, raw, `"unresolved_filters":[]`)

		var result batchOutput
		require.NoError(t, json.NewDecoder(strings.NewReader(raw)).Decode(&result))
		require.Len(t, result.Deletes, 1)
		assert.Equal(t, "viewer", result.Deletes[0].Relation)
	})

	t.Run("warns on stderr for unresolved filters in JSONL", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/with_filter.yaml", testOpts("jsonl", false), strings.NewReader(`{"id":"anne","org":"acme"}`), &out, &errOut)
		require.NoError(t, err)

		assert.Contains(t, errOut.String(), "unresolved tuple filter")
		assert.Contains(t, errOut.String(), "action=delete")
		assert.Contains(t, errOut.String(), "user=user:anne")
		assert.Contains(t, errOut.String(), "object=org:acme")
	})

	t.Run("--input reads from file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		inputFile := filepath.Join(dir, "event.json")
		require.NoError(t, os.WriteFile(inputFile, []byte(`{"id":"anne"}`), 0o600))

		inputFd, err := os.Open(inputFile)
		require.NoError(t, err)

		defer inputFd.Close()

		var out bytes.Buffer

		err = runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false), inputFd, &out, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Contains(t, out.String(), "user:anne")
	})

	t.Run("missing mapping file returns error", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/nonexistent.yaml", testOpts("jsonl", false), strings.NewReader(`{}`), &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
	})

	t.Run("invalid JSON input returns error", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false), strings.NewReader(`not json`), &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input JSON")
	})

	t.Run("null input is rejected as a non-object record", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false), strings.NewReader(`null`), &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.ErrorIs(t, err, errNonObjectRecord)
	})

	t.Run("invalid mapping returns errMappingInvalid", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/invalid.yaml", testOpts("jsonl", false), strings.NewReader(`{}`), &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.ErrorIs(t, err, errMappingInvalid)
	})

	t.Run("unknown format returns errUnknownRunFormat", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("xml", false), strings.NewReader(`{}`), &bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.ErrorIs(t, err, errUnknownRunFormat)
	})
}

func TestRunMappingMultiRecord(t *testing.T) {
	t.Parallel()

	t.Run("multi-record JSONL streams one op per record", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false),
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"bob"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)

		users := make([]string, 0, len(lines))

		for _, line := range lines {
			var op opTupleOutput
			require.NoError(t, json.Unmarshal([]byte(line), &op))
			users = append(users, op.User)
		}

		assert.ElementsMatch(t, []string{"user:anne", "user:bob"}, users)
	})

	t.Run("multi-record --format json unions all records into one batch", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("json", false),
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"bob"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.Writes, 2)
		assert.ElementsMatch(t, []string{"user:anne", "user:bob"},
			[]string{result.Writes[0].User, result.Writes[1].User})
	})

	t.Run("multi-record --format json --writes-only is one flat array across records", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("json", true),
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"bob"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var writes []writeTupleOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&writes))
		require.Len(t, writes, 2)
	})

	t.Run("--aggregate dedups an identical tuple produced across records", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "jsonl", aggregate: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		assert.Len(t, lines, 1)
	})

	t.Run("--aggregate dedups identical unresolved filters across records", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/with_filter.yaml",
			runMappingOptions{format: "json", aggregate: true},
			strings.NewReader(`{"id":"anne","org":"acme"}`+"\n"+`{"id":"anne","org":"acme"}`), &out, &errOut)
		require.NoError(t, err)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.UnresolvedFilters, 1)
	})

	t.Run("--aggregate --format json dedups across records", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "json", aggregate: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"anne"}`), &out, &bytes.Buffer{})
		require.NoError(t, err)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.Writes, 1)
	})

	t.Run("--aggregate surfaces a cross-record write/delete conflict as a runtime error", func(t *testing.T) {
		t.Parallel()

		err := runMapping(context.Background(), "testdata/conflicting.yaml",
			runMappingOptions{format: "jsonl", aggregate: true},
			strings.NewReader(`{"id":"anne","org":"acme","op":"add"}`+"\n"+`{"id":"anne","org":"acme","op":"remove"}`),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "conflict")
		require.NotErrorIs(t, err, errMappingInvalid)
		require.NotErrorIs(t, err, errUnknownRunFormat)
	})
}

func TestRunMappingContinueOnError(t *testing.T) {
	t.Parallel()

	t.Run("a malformed record is fatal without --continue-on-error", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml", testOpts("jsonl", false),
			strings.NewReader(`{"id":"anne"}`+"\n"+`not json`+"\n"+`{"id":"bob"}`), &out, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input JSON")
		require.NotErrorIs(t, err, errRecordsFailed)
	})

	t.Run("--continue-on-error skips a malformed record and emits the rest", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "jsonl", continueOnError: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`not json`+"\n"+`{"id":"bob"}`), &out, &errOut)

		require.ErrorIs(t, err, errRecordsFailed)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)

		users := make([]string, 0, len(lines))

		for _, line := range lines {
			var op opTupleOutput
			require.NoError(t, json.Unmarshal([]byte(line), &op))
			users = append(users, op.User)
		}

		assert.ElementsMatch(t, []string{"user:anne", "user:bob"}, users)
		assert.Contains(t, errOut.String(), "line 2")
	})

	t.Run("--continue-on-error skips a record that fails evaluation", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		// valid.yaml interpolates input.id; a record without it fails evaluation.
		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "jsonl", continueOnError: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"name":"nobody"}`+"\n"+`{"id":"bob"}`), &out, &errOut)

		require.ErrorIs(t, err, errRecordsFailed)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)
		assert.Contains(t, errOut.String(), "line 2")
	})

	t.Run("--continue-on-error with all records valid returns no error", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "jsonl", continueOnError: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`{"id":"bob"}`), &out, &errOut)

		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)
	})

	t.Run("--continue-on-error reports the physical line number past a blank line", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		// A blank line counts as a physical line, so the malformed record is line 3.
		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "jsonl", continueOnError: true},
			strings.NewReader(`{"id":"anne"}`+"\n\n"+`not json`), &out, &errOut)

		require.ErrorIs(t, err, errRecordsFailed)
		assert.Contains(t, errOut.String(), "line 3")

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 1)
	})

	t.Run("--continue-on-error aggregates only the surviving records", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMapping(context.Background(), "testdata/valid.yaml",
			runMappingOptions{format: "json", aggregate: true, continueOnError: true},
			strings.NewReader(`{"id":"anne"}`+"\n"+`not json`+"\n"+`{"id":"anne"}`), &out, &errOut)

		require.ErrorIs(t, err, errRecordsFailed)

		var result batchOutput
		require.NoError(t, json.NewDecoder(&out).Decode(&result))
		require.Len(t, result.Writes, 1)
	})
}
