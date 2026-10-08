package mapping

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	openfga "github.com/openfga/go-sdk"
	sdkclient "github.com/openfga/go-sdk/client"
	"github.com/openfga/mapper/language"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReadRequest implements sdkclient.SdkClientReadRequestInterface.
// It captures the Options call so tests can assert consistency preference.
type mockReadRequest struct {
	capturedOptions sdkclient.ClientReadOptions
	tuples          []openfga.Tuple
	err             error
}

func (m *mockReadRequest) Options(opts sdkclient.ClientReadOptions) sdkclient.SdkClientReadRequestInterface {
	m.capturedOptions = opts
	return m
}

func (m *mockReadRequest) Body(_ sdkclient.ClientReadRequest) sdkclient.SdkClientReadRequestInterface {
	return m
}

func (m *mockReadRequest) Execute() (*openfga.ReadResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &openfga.ReadResponse{Tuples: m.tuples}, nil
}

func (m *mockReadRequest) GetStoreIdOverride() *string              { return nil } //nolint:staticcheck
func (m *mockReadRequest) GetContext() context.Context              { return context.Background() }
func (m *mockReadRequest) GetBody() *sdkclient.ClientReadRequest    { return nil }
func (m *mockReadRequest) GetOptions() *sdkclient.ClientReadOptions { return nil }

// mockWriteRequest implements sdkclient.SdkClientWriteRequestInterface.
// It captures the Body and returns a configured response from Execute.
type mockWriteRequest struct {
	capturedBody sdkclient.ClientWriteRequest
	response     *sdkclient.ClientWriteResponse
	err          error
}

func (m *mockWriteRequest) Options(_ sdkclient.ClientWriteOptions) sdkclient.SdkClientWriteRequestInterface {
	return m
}

func (m *mockWriteRequest) Body(b sdkclient.ClientWriteRequest) sdkclient.SdkClientWriteRequestInterface {
	m.capturedBody = b
	return m
}

func (m *mockWriteRequest) Execute() (*sdkclient.ClientWriteResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &sdkclient.ClientWriteResponse{}, nil
}

func (m *mockWriteRequest) GetAuthorizationModelIdOverride() *string  { return nil } //nolint:staticcheck
func (m *mockWriteRequest) GetStoreIdOverride() *string               { return nil } //nolint:staticcheck
func (m *mockWriteRequest) GetContext() context.Context               { return context.Background() }
func (m *mockWriteRequest) GetOptions() *sdkclient.ClientWriteOptions { return nil }
func (m *mockWriteRequest) GetBody() *sdkclient.ClientWriteRequest    { return nil }

// mockSdkClient implements sdkWriteReadClient.
// writeReqFn, when set, is called on each Write() invocation; otherwise writeReq is returned.
type mockSdkClient struct {
	readReq    *mockReadRequest
	writeReq   *mockWriteRequest
	writeReqFn func() sdkclient.SdkClientWriteRequestInterface
}

func (m *mockSdkClient) Read(_ context.Context) sdkclient.SdkClientReadRequestInterface {
	return m.readReq
}

func (m *mockSdkClient) Write(_ context.Context) sdkclient.SdkClientWriteRequestInterface {
	if m.writeReqFn != nil {
		return m.writeReqFn()
	}
	return m.writeReq
}

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

func TestFGATupleClientReadTuples(t *testing.T) {
	t.Parallel()

	t.Run("passes HIGHER_CONSISTENCY and filter to Read", func(t *testing.T) {
		t.Parallel()

		existing := openfga.Tuple{Key: openfga.TupleKey{
			User:     "user:anne",
			Relation: "admin",
			Object:   "org:acme",
		}}
		readReq := &mockReadRequest{tuples: []openfga.Tuple{existing}}
		fc := &fgaTupleClient{inner: &mockSdkClient{readReq: readReq}}

		filter := language.TupleFilter{User: "user:anne", Object: "org:acme"}
		tuples, err := fc.ReadTuples(context.Background(), filter)

		require.NoError(t, err)
		require.Len(t, tuples, 1)
		assert.Equal(t, "user:anne", tuples[0].User)
		assert.Equal(t, "admin", tuples[0].Relation)
		require.NotNil(t, readReq.capturedOptions.Consistency)
		assert.Equal(t, openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY, *readReq.capturedOptions.Consistency)
	})
}

func TestFGATupleClientWriteTuples(t *testing.T) {
	t.Parallel()

	t.Run("per-tuple FAILURE on delete propagates as error", func(t *testing.T) {
		t.Parallel()

		writeReq := &mockWriteRequest{
			response: &sdkclient.ClientWriteResponse{
				Deletes: []sdkclient.ClientWriteRequestDeleteResponse{
					{
						TupleKey: sdkclient.ClientTupleKeyWithoutCondition{
							User:     "user:anne",
							Relation: "member",
							Object:   "org:acme",
						},
						Status: sdkclient.FAILURE,
					},
				},
			},
		}
		fc := &fgaTupleClient{inner: &mockSdkClient{writeReq: writeReq}}

		err := fc.WriteTuples(context.Background(), []language.Tuple{
			{User: "user:anne", Relation: "member", Object: "org:acme", Action: language.ActionDelete},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting tuples")
	})

	t.Run("per-tuple FAILURE on delete prevents writes", func(t *testing.T) {
		t.Parallel()

		var writeCalls int
		sdk := &mockSdkClient{
			writeReqFn: func() sdkclient.SdkClientWriteRequestInterface {
				writeCalls++
				if writeCalls > 1 {
					t.Error("Write must not be called after a delete failure")
				}
				return &mockWriteRequest{
					response: &sdkclient.ClientWriteResponse{
						Deletes: []sdkclient.ClientWriteRequestDeleteResponse{
							{
								Status:   sdkclient.FAILURE,
								TupleKey: sdkclient.ClientTupleKeyWithoutCondition{User: "user:anne", Relation: "admin", Object: "org:acme"},
							},
						},
					},
				}
			},
		}
		fc := &fgaTupleClient{inner: sdk}

		err := fc.WriteTuples(context.Background(), []language.Tuple{
			{User: "user:anne", Relation: "admin", Object: "org:acme", Action: language.ActionDelete},
			{User: "user:anne", Relation: "viewer", Object: "org:acme", Action: language.ActionWrite},
		})

		require.Error(t, err)
		assert.Equal(t, 1, writeCalls, "Write must not be called a second time after delete failure")
	})
}
