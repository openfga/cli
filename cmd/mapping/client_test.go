package mapping

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openfga/mapper/language"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type alwaysErrWriter struct{}

func (e *alwaysErrWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestCountingClient(t *testing.T) {
	t.Parallel()

	t.Run("counts writes and deletes", func(t *testing.T) {
		t.Parallel()

		cc := &countingClient{inner: &mockTupleClient{}}
		tuples := []language.Tuple{
			{User: "user:anne", Relation: "member", Object: "org:acme", Action: language.ActionWrite},
			{User: "user:bob", Relation: "member", Object: "org:acme", Action: language.ActionWrite},
			{User: "user:eve", Relation: "viewer", Object: "org:acme", Action: language.ActionDelete},
		}

		require.NoError(t, cc.WriteTuples(context.Background(), tuples))
		assert.Equal(t, 2, cc.writes)
		assert.Equal(t, 1, cc.dels)
	})

	t.Run("delegates WriteTuples to inner", func(t *testing.T) {
		t.Parallel()

		var received []language.Tuple
		inner := &mockTupleClient{
			writeFn: func(_ context.Context, tuples []language.Tuple) error {
				received = tuples
				return nil
			},
		}
		cc := &countingClient{inner: inner}
		tuples := []language.Tuple{
			{User: "user:anne", Relation: "member", Object: "org:acme"},
		}

		require.NoError(t, cc.WriteTuples(context.Background(), tuples))
		assert.Equal(t, tuples, received)
	})

	t.Run("delegates ReadTuples to inner", func(t *testing.T) {
		t.Parallel()

		expected := []language.Tuple{{User: "user:anne", Relation: "member", Object: "org:acme"}}
		inner := &mockTupleClient{
			readFn: func(_ context.Context, _ language.TupleFilter) ([]language.Tuple, error) {
				return expected, nil
			},
		}
		cc := &countingClient{inner: inner}

		got, err := cc.ReadTuples(context.Background(), language.TupleFilter{})
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}

func TestPrintOnlyClient(t *testing.T) {
	t.Parallel()

	t.Run("WriteTuples encodes tuples as JSONL with op field", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		pc := &printOnlyClient{inner: &mockTupleClient{}, out: &out}
		tuples := []language.Tuple{
			{User: "user:anne", Relation: "member", Object: "org:acme", Action: language.ActionWrite},
			{User: "user:bob", Relation: "viewer", Object: "org:acme", Action: language.ActionDelete},
		}

		require.NoError(t, pc.WriteTuples(context.Background(), tuples))

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		require.Len(t, lines, 2)
		assert.Contains(t, lines[0], `"user:anne"`)
		assert.Contains(t, lines[0], `"op"`)
		assert.Contains(t, lines[1], `"user:bob"`)
	})

	t.Run("WriteTuples returns error when writer fails", func(t *testing.T) {
		t.Parallel()

		pc := &printOnlyClient{inner: &mockTupleClient{}, out: &alwaysErrWriter{}}
		tuples := []language.Tuple{
			{User: "user:anne", Relation: "member", Object: "org:acme"},
		}

		err := pc.WriteTuples(context.Background(), tuples)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "encoding dry-run output")
	})

	t.Run("delegates ReadTuples to inner", func(t *testing.T) {
		t.Parallel()

		expected := []language.Tuple{{User: "user:anne", Relation: "member", Object: "org:acme"}}
		inner := &mockTupleClient{
			readFn: func(_ context.Context, _ language.TupleFilter) ([]language.Tuple, error) {
				return expected, nil
			},
		}
		pc := &printOnlyClient{inner: inner, out: &bytes.Buffer{}}

		got, err := pc.ReadTuples(context.Background(), language.TupleFilter{})
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}

func TestPartitionSyncTuples(t *testing.T) {
	t.Parallel()

	t.Run("write tuple goes to writes slice", func(t *testing.T) {
		t.Parallel()

		tuple := language.Tuple{
			User:     "user:anne",
			Relation: "member",
			Object:   "org:acme",
			Action:   language.ActionWrite,
		}
		writes, deletes := partitionSyncTuples([]language.Tuple{tuple})
		require.Len(t, writes, 1)
		assert.Empty(t, deletes)
		assert.Equal(t, "user:anne", writes[0].User)
		assert.Equal(t, "member", writes[0].Relation)
		assert.Equal(t, "org:acme", writes[0].Object)
	})

	t.Run("delete tuple goes to deletes slice", func(t *testing.T) {
		t.Parallel()

		tuple := language.Tuple{
			User:     "user:anne",
			Relation: "member",
			Object:   "org:acme",
			Action:   language.ActionDelete,
		}
		writes, deletes := partitionSyncTuples([]language.Tuple{tuple})
		assert.Empty(t, writes)
		require.Len(t, deletes, 1)
		assert.Equal(t, "user:anne", deletes[0].User)
		assert.Equal(t, "member", deletes[0].Relation)
		assert.Equal(t, "org:acme", deletes[0].Object)
	})

	t.Run("write tuple with condition sets Condition on write key", func(t *testing.T) {
		t.Parallel()

		tuple := language.Tuple{
			User:      "user:anne",
			Relation:  "owner",
			Object:    "group:eng",
			Condition: "check_ip",
		}
		writes, _ := partitionSyncTuples([]language.Tuple{tuple})
		require.Len(t, writes, 1)
		require.NotNil(t, writes[0].Condition)
		assert.Equal(t, "check_ip", writes[0].Condition.Name)
		assert.Nil(t, writes[0].Condition.Context)
	})

	t.Run("write tuple with condition and context sets Context on write key", func(t *testing.T) {
		t.Parallel()

		ctx := map[string]any{"ip_addr": "192.168.1.1"}
		tuple := language.Tuple{
			User:      "user:anne",
			Relation:  "owner",
			Object:    "group:eng",
			Condition: "check_ip",
			Context:   ctx,
		}
		writes, _ := partitionSyncTuples([]language.Tuple{tuple})
		require.Len(t, writes, 1)
		require.NotNil(t, writes[0].Condition)
		require.NotNil(t, writes[0].Condition.Context)
		assert.Equal(t, ctx, *writes[0].Condition.Context)
	})

	t.Run("empty input returns nil slices", func(t *testing.T) {
		t.Parallel()

		writes, deletes := partitionSyncTuples(nil)
		assert.Nil(t, writes)
		assert.Nil(t, deletes)
	})
}
