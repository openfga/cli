package mapping

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/openfga/mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runREPL(t *testing.T, path, input string) (string, string) {
	t.Helper()

	var out, errOut bytes.Buffer

	err := runMappingInteractive(context.Background(), path, strings.NewReader(input), &out, &errOut)
	require.NoError(t, err)

	return out.String(), errOut.String()
}

func TestRunMappingInteractive(t *testing.T) {
	t.Parallel()

	t.Run("banner reports the rule count and command help", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", ":quit\n")
		assert.Contains(t, out, "mapping loaded: 1 rules")
		assert.Contains(t, out, ":reload")
		assert.Contains(t, out, ":trace on|off")
		assert.Contains(t, out, ":quit")
	})

	t.Run("evaluates a pasted document ended by a blank line", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", `{"id":"anne"}`+"\n\n:quit\n")
		assert.Contains(t, out, "write")
		assert.Contains(t, out, "user:anne")
		assert.Contains(t, out, "member")
		assert.Contains(t, out, "org:acme")
	})

	t.Run("joins a multi-line document before evaluating", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", "{\n\"id\":\"anne\"}\n\n:quit\n")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("evaluates a single-line document on its own Enter without a blank line or quit", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", `{"id":"anne"}`+"\n")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("assembles a multi-line document and evaluates it once complete", func(t *testing.T) {
		t.Parallel()

		// No blank-line terminator and no :quit: the document is evaluated as
		// soon as the accumulated lines form a complete JSON value.
		out, _ := runREPL(t, "testdata/valid.yaml", "{\n  \"id\": \"anne\"\n}\n")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("renders a delete operation", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/mixed_actions.yaml", `{"id":"anne","org":"acme"}`+"\n\n:quit\n")
		assert.Contains(t, out, "delete")
		assert.Contains(t, out, "viewer")
	})

	t.Run("renders an unresolved tuple filter", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/with_filter.yaml", `{"id":"anne","org":"acme"}`+"\n\n:quit\n")
		assert.Contains(t, out, "filter")
		assert.Contains(t, out, "org:acme")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("renders a filter operation's desired-state tuples", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/filter_with_desired.yaml", `{"id":"anne","org":"acme"}`+"\n\n:quit\n")
		assert.Contains(t, out, "filter")
		assert.Contains(t, out, "desired")
		assert.Contains(t, out, "viewer")
		// The desired-state tuple the filter reconciles toward.
		assert.Regexp(t, `desired\s+user:anne\s+viewer\s+org:acme`, out)
	})

	t.Run("reports no tuples when a guarded rule is skipped", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/guarded.yaml", `{"type":"nope","id":"anne"}`+"\n\n:quit\n")
		assert.Contains(t, out, "(no tuples)")
	})

	t.Run(":trace on shows the per-rule summary", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", ":trace on\n"+`{"id":"anne"}`+"\n\n:quit\n")
		assert.Contains(t, out, "rules matched")
		assert.Contains(t, out, "members")
	})

	t.Run(":trace on reports skipped rules", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/guarded.yaml", ":trace on\n"+`{"type":"nope"}`+"\n\n:quit\n")
		assert.Contains(t, out, "rules skipped")
		assert.Contains(t, out, "guarded")
	})

	t.Run("a malformed document warns and the loop continues", func(t *testing.T) {
		t.Parallel()

		out, errOut := runREPL(t, "testdata/valid.yaml", "not json\n\n"+`{"id":"anne"}`+"\n\n:quit\n")
		assert.Contains(t, errOut, "JSON")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("invalid JSON reports the line, column, and a caret", func(t *testing.T) {
		t.Parallel()

		_, errOut := runREPL(t, "testdata/valid.yaml", `{"id": bob}`+"\n")
		assert.Contains(t, errOut, "line 1, column 8")
		assert.Contains(t, errOut, `{"id": bob}`)
		assert.Contains(t, errOut, "       ^")
	})

	t.Run("an unterminated string evaluates and errors rather than waiting", func(t *testing.T) {
		t.Parallel()

		// `{"id}` is incomplete only because the string never closes; a newline
		// can never make it valid, so it is evaluated on Enter, not continued.
		_, errOut := runREPL(t, "testdata/valid.yaml", `{"id}`+"\n")
		assert.Contains(t, errOut, "invalid JSON")
		assert.NotContains(t, errOut, "incomplete JSON")
	})

	t.Run(":reload recompiles the mapping from disk", func(t *testing.T) {
		t.Parallel()

		out, _ := runREPL(t, "testdata/valid.yaml", ":reload\n:quit\n")
		assert.Contains(t, out, "mapping reloaded: 1 rules")
	})

	t.Run("a blank command is a no-op and does not panic", func(t *testing.T) {
		t.Parallel()

		session := &interactiveSession{out: &bytes.Buffer{}, errOut: &bytes.Buffer{}}
		assert.False(t, session.handleCommand(""))
		assert.False(t, session.handleCommand("   "))
	})

	t.Run("an unknown command warns and the loop continues", func(t *testing.T) {
		t.Parallel()

		out, errOut := runREPL(t, "testdata/valid.yaml", ":bogus\n"+`{"id":"anne"}`+"\n\n:quit\n")
		assert.Contains(t, errOut, "unknown command")
		assert.Contains(t, out, "user:anne")
	})

	t.Run("invalid mapping returns errMappingInvalid", func(t *testing.T) {
		t.Parallel()

		var out, errOut bytes.Buffer

		err := runMappingInteractive(context.Background(), "testdata/invalid.yaml", strings.NewReader(":quit\n"), &out, &errOut)
		require.ErrorIs(t, err, errMappingInvalid)
	})
}

func TestClassifyJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want jsonCompleteness
	}{
		{"complete object", `{"id":"anne"}`, jsonComplete},
		{"complete across lines", "{\n\"id\":\"anne\"\n}", jsonComplete},
		{"open brace only", "{", jsonIncomplete},
		{"partial object", `{"id":`, jsonIncomplete},
		{"unterminated string", `{"id":"an`, jsonIncomplete},
		{"not json", "not json", jsonInvalid},
		{"leading garbage", "xyz{", jsonInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, classifyJSON(tc.in))
		})
	}
}

func TestInteractiveIncompleteNudge(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/valid.yaml")
	require.NoError(t, err)

	compiled, err := mapper.Compile(data, mapper.WithTrace(true))
	require.NoError(t, err)

	var out, errOut bytes.Buffer

	session := &interactiveSession{
		ctx:         context.Background(),
		path:        "testdata/valid.yaml",
		compiled:    compiled,
		out:         &out,
		errOut:      &errOut,
		interactive: true,
	}

	// "{" is incomplete; the blank line then forces evaluation, surfacing the
	// end-of-input error with a position.
	reader := &scannerLineReader{
		scanner: bufio.NewScanner(strings.NewReader("{\n\n")),
		out:     &out,
		prompt:  promptPrimary,
	}
	require.NoError(t, session.loop(reader))

	assert.Contains(t, errOut.String(), "incomplete JSON")
	assert.Contains(t, errOut.String(), "unexpected end of JSON input")
}

func TestInteractiveNudgeSuppressedWhenNotInteractive(t *testing.T) {
	t.Parallel()

	// The plain (piped/test) path leaves interactive false, so no nudge noise.
	_, errOut := runREPL(t, "testdata/valid.yaml", "{\n\n:quit\n")
	assert.NotContains(t, errOut, "incomplete JSON")
}

func TestLocateOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		doc        string
		offset     int
		wantLine   int
		wantColumn int
		wantText   string
	}{
		{"single line", `{"id": bob}`, 8, 1, 8, `{"id": bob}`},
		{"multi line", "{\n  \"id\": bob\n}", 11, 2, 9, `  "id": bob`},
		{"end of input", `{"id":"anne"`, 12, 1, 12, `{"id":"anne"`},
		{"offset past end clamps", "{", 5, 1, 1, "{"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			line, column, text := locateOffset(testCase.doc, testCase.offset)
			assert.Equal(t, testCase.wantLine, line)
			assert.Equal(t, testCase.wantColumn, column)
			assert.Equal(t, testCase.wantText, text)
		})
	}
}

func TestAwaitingMoreInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"open brace", "{", true},
		{"after colon", `{"id":`, true},
		{"after comma", "[1,", true},
		{"unclosed object", `{"id":"anne"`, true},
		{"complete object", `{"id":"anne"}`, false},
		{"unterminated string", `{"id}`, false},
		{"unterminated value string", `{"a":"b`, false},
		{"partial literal", "tru", false},
		{"not json", "not json", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.want, awaitingMoreInput(testCase.in))
		})
	}
}

func TestCheckInteractiveFlags(t *testing.T) {
	t.Parallel()

	t.Run("rejects --writes-only", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, checkInteractiveFlags(true, ""), errInteractiveWithWritesOnly)
	})

	t.Run("rejects --input", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, checkInteractiveFlags(false, "event.json"), errInteractiveWithInput)
	})

	t.Run("accepts a clean interactive invocation", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, checkInteractiveFlags(false, ""))
	})
}
