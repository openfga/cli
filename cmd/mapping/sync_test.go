package mapping

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/openfga/mapper/apply"
	"github.com/openfga/mapper/language"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTupleClient struct {
	readFn  func(ctx context.Context, filter language.TupleFilter) ([]language.Tuple, error)
	writeFn func(ctx context.Context, tuples []language.Tuple) error
}

func (m *mockTupleClient) ReadTuples(ctx context.Context, filter language.TupleFilter) ([]language.Tuple, error) {
	if m.readFn != nil {
		return m.readFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockTupleClient) WriteTuples(ctx context.Context, tuples []language.Tuple) error {
	if m.writeFn != nil {
		return m.writeFn(ctx, tuples)
	}
	return nil
}

func TestSyncMapping(t *testing.T) {
	t.Parallel()

	t.Run("happy path processes two records", func(t *testing.T) {
		t.Parallel()

		var called int
		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				called++
				return nil
			},
		}

		var out, errOut bytes.Buffer
		err := syncMapping(context.Background(), "testdata/valid.yaml", syncOptions{},
			client, strings.NewReader("{\"id\":\"anne\"}\n{\"id\":\"bob\"}\n"),
			&out, &errOut)
		require.NoError(t, err)
		assert.Equal(t, 2, called)
		assert.Empty(t, out.String())
		assert.Contains(t, errOut.String(), "✓")
	})

	t.Run("quiet suppresses all progress output", func(t *testing.T) {
		t.Parallel()

		var errOut bytes.Buffer
		err := syncMapping(context.Background(), "testdata/valid.yaml", syncOptions{quiet: true},
			&mockTupleClient{}, strings.NewReader("{\"id\":\"anne\"}\n"),
			&bytes.Buffer{}, &errOut)
		require.NoError(t, err)
		assert.Empty(t, errOut.String())
	})

	t.Run("invalid mapping returns errMappingInvalid and writes diagnostics to errOut", func(t *testing.T) {
		t.Parallel()

		var errOut bytes.Buffer
		err := syncMapping(context.Background(), "testdata/invalid.yaml", syncOptions{},
			&mockTupleClient{}, strings.NewReader(""),
			&bytes.Buffer{}, &errOut)
		require.ErrorIs(t, err, errMappingInvalid)
		assert.NotEmpty(t, errOut.String())
	})

	t.Run("missing mapping file returns error containing 'reading'", func(t *testing.T) {
		t.Parallel()

		err := syncMapping(context.Background(), "testdata/nonexistent.yaml", syncOptions{},
			&mockTupleClient{}, strings.NewReader(""),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading")
	})

	t.Run("malformed JSON record returns error without continue-on-error", func(t *testing.T) {
		t.Parallel()

		err := syncMapping(context.Background(), "testdata/valid.yaml", syncOptions{},
			&mockTupleClient{}, strings.NewReader("not json\n"),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.False(t, errors.Is(err, errSyncRecordsFailed))
		assert.False(t, errors.Is(err, errMappingInvalid))
	})

	t.Run("continue-on-error skips bad records and processes good ones", func(t *testing.T) {
		t.Parallel()

		var called int
		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				called++
				return nil
			},
		}

		var errOut bytes.Buffer
		err := syncMapping(context.Background(), "testdata/valid.yaml",
			syncOptions{continueOnError: true},
			client,
			strings.NewReader("not json\n{\"id\":\"anne\"}\n"),
			&bytes.Buffer{}, &errOut)
		require.ErrorIs(t, err, errSyncRecordsFailed)
		assert.Equal(t, 1, called)
		assert.Contains(t, errOut.String(), "✗")
	})

	t.Run("continue-on-error with all valid records returns nil", func(t *testing.T) {
		t.Parallel()

		var called int
		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				called++
				return nil
			},
		}

		err := syncMapping(context.Background(), "testdata/valid.yaml",
			syncOptions{continueOnError: true},
			client,
			strings.NewReader("{\"id\":\"anne\"}\n{\"id\":\"bob\"}\n"),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.NoError(t, err)
		assert.Equal(t, 2, called)
	})

	t.Run("dry-run does not call WriteTuples on the base client", func(t *testing.T) {
		t.Parallel()

		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				t.Fatal("WriteTuples must not be called in dry-run")
				return nil
			},
		}

		err := syncMapping(context.Background(), "testdata/valid.yaml",
			syncOptions{dryRun: true},
			client,
			strings.NewReader("{\"id\":\"anne\"}\n"),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.NoError(t, err)
	})

	t.Run("dry-run emits JSONL to stdout when stdout is a non-TTY file", func(t *testing.T) {
		t.Parallel()

		pr, pw, err := os.Pipe()
		require.NoError(t, err)
		defer pr.Close()

		syncErr := syncMapping(context.Background(), "testdata/valid.yaml",
			syncOptions{dryRun: true},
			&mockTupleClient{},
			strings.NewReader("{\"id\":\"anne\"}\n"),
			pw, &bytes.Buffer{})
		pw.Close()
		require.NoError(t, syncErr)

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, pr)
		assert.Contains(t, buf.String(), "user:anne")
	})

	t.Run("error without continue-on-error does not write subsequent records", func(t *testing.T) {
		t.Parallel()

		var called int
		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				called++
				return nil
			},
		}

		err := syncMapping(context.Background(), "testdata/valid.yaml", syncOptions{},
			client,
			strings.NewReader("not json\n{\"id\":\"anne\"}\n"),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		assert.Equal(t, 0, called, "valid record after failure must not be written")
	})

	t.Run("WriteTuples error is returned wrapped in WriteError", func(t *testing.T) {
		t.Parallel()

		writeErr := errors.New("store unavailable")
		client := &mockTupleClient{
			writeFn: func(_ context.Context, _ []language.Tuple) error {
				return writeErr
			},
		}

		err := syncMapping(context.Background(), "testdata/valid.yaml", syncOptions{},
			client,
			strings.NewReader("{\"id\":\"anne\"}\n"),
			&bytes.Buffer{}, &bytes.Buffer{})
		require.Error(t, err)
		var wErr *apply.WriteError
		assert.True(t, errors.As(err, &wErr))
	})
}
