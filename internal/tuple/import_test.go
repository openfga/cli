package tuple

import (
	"context"
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
)

func TestImportTuplesRejectsValuesOutsideInt32Range(t *testing.T) {
	t.Parallel()

	if strconv.IntSize == 32 {
		t.Skip("int cannot exceed math.MaxInt32 on 32-bit platforms")
	}

	maxInt32 := int(math.MaxInt32)
	maxInt32PlusOne := maxInt32
	maxInt32PlusOne++
	maxInt32String := strconv.Itoa(maxInt32)
	tests := []struct {
		name                string
		maxTuplesPerWrite   int
		maxParallelRequests int
		expectedError       string
	}{
		{
			name:                "max tuples per write",
			maxTuplesPerWrite:   maxInt32PlusOne,
			maxParallelRequests: 1,
			expectedError:       "maxTuplesPerWrite must be at most " + maxInt32String,
		},
		{
			name:                "max parallel requests",
			maxTuplesPerWrite:   1,
			maxParallelRequests: maxInt32PlusOne,
			expectedError:       "maxParallelRequests must be at most " + maxInt32String,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := ImportTuples(
				context.Background(), nil, 0, 0, 0, test.maxTuplesPerWrite, test.maxParallelRequests,
				client.ClientWriteRequest{}, client.ClientWriteOptions{},
			)

			assert.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestProcessWrites(t *testing.T) {
	t.Parallel()

	writes := []client.ClientWriteRequestWriteResponse{
		{
			TupleKey: client.ClientTupleKey{User: "user:1", Relation: "access", Object: "document:1"},
			Status:   client.SUCCESS,
		},
		{
			TupleKey: client.ClientTupleKey{User: "user:2", Relation: "access", Object: "document:2"},
			Status:   client.FAILURE,
			Error:    errors.New("error message: some error"),
		},
	}

	successful, failed := processWrites(writes)

	assert.Len(t, successful, 1)
	assert.Len(t, failed, 1)
	assert.Equal(t, "error message: some error", failed[0].Reason)
}

func TestProcessDeletes(t *testing.T) {
	t.Parallel()

	deletes := []client.ClientWriteRequestDeleteResponse{
		{
			TupleKey: client.ClientTupleKeyWithoutCondition{User: "user:1", Relation: "access", Object: "document:1"},
			Status:   client.SUCCESS,
		},
		{
			TupleKey: client.ClientTupleKeyWithoutCondition{User: "user:2", Relation: "access", Object: "document:2"},
			Status:   client.FAILURE,
			Error:    errors.New("error message: some error"),
		},
	}

	successful, failed := processDeletes(deletes)

	assert.Len(t, successful, 1)
	assert.Len(t, failed, 1)
	assert.Equal(t, "error message: some error", failed[0].Reason)
}

func TestGetImportChunk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		index             int
		maxTuplesPerWrite int
		writes            []client.ClientTupleKey
		deletes           []client.ClientTupleKeyWithoutCondition
		expectedWrites    []client.ClientTupleKey
		expectedDeletes   []client.ClientTupleKeyWithoutCondition
	}{
		{
			name:              "Empty writes and deletes",
			index:             0,
			maxTuplesPerWrite: 2,
			writes:            []client.ClientTupleKey{},
			deletes:           []client.ClientTupleKeyWithoutCondition{},
			expectedWrites:    []client.ClientTupleKey{},
			expectedDeletes:   []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Only writes, within limit",
			index:             0,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			deletes:         []client.ClientTupleKeyWithoutCondition{},
			expectedWrites:  []client.ClientTupleKey{{User: "user:1", Relation: "access", Object: "document:1"}, {User: "user:2", Relation: "access", Object: "document:2"}},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Only writes, exceeding limit",
			index:             0,
			maxTuplesPerWrite: 1,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			deletes:         []client.ClientTupleKeyWithoutCondition{},
			expectedWrites:  []client.ClientTupleKey{{User: "user:1", Relation: "access", Object: "document:1"}},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Index out of range for writes",
			index:             1,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedWrites:  []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Only deletes, within limit",
			index:             0,
			maxTuplesPerWrite: 2,
			writes:            []client.ClientTupleKey{},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedWrites:  []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{{User: "user:1", Relation: "access", Object: "document:1"}, {User: "user:2", Relation: "access", Object: "document:2"}},
		},
		{
			name:              "Only deletes, exceeding limit",
			index:             0,
			maxTuplesPerWrite: 1,
			writes:            []client.ClientTupleKey{},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedWrites:  []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{{User: "user:1", Relation: "access", Object: "document:1"}},
		},
		{
			name:              "Only deletes, exceeding limit, index = 1",
			index:             1,
			maxTuplesPerWrite: 2,
			writes:            []client.ClientTupleKey{},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
				{User: "user:5", Relation: "access", Object: "document:5"},
			},
			expectedWrites:  []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{{User: "user:3", Relation: "access", Object: "document:3"}, {User: "user:4", Relation: "access", Object: "document:4"}},
		},
		{
			name:              "Index out of range for deletes",
			index:             1,
			maxTuplesPerWrite: 2,
			writes:            []client.ClientTupleKey{},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:1", Relation: "access", Object: "document:1"},
			},
			expectedWrites:  []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Mixed writes and deletes, within limit",
			index:             0,
			maxTuplesPerWrite: 3,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:2", Relation: "access", Object: "document:2"},
				{User: "user:3", Relation: "access", Object: "document:3"},
			},
			expectedWrites:  []client.ClientTupleKey{{User: "user:1", Relation: "access", Object: "document:1"}},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{{User: "user:2", Relation: "access", Object: "document:2"}, {User: "user:3", Relation: "access", Object: "document:3"}},
		},
		{
			name:              "Mixed writes and deletes, exceeding limit",
			index:             0,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:2", Relation: "access", Object: "document:2"},
				{User: "user:3", Relation: "access", Object: "document:3"},
			},
			expectedWrites:  []client.ClientTupleKey{{User: "user:1", Relation: "access", Object: "document:1"}},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{{User: "user:2", Relation: "access", Object: "document:2"}},
		},
		{
			name:              "Mixed writes and deletes, exceeding limit",
			index:             0,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
				{User: "user:3", Relation: "access", Object: "document:3"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:4", Relation: "access", Object: "document:4"},
				{User: "user:5", Relation: "access", Object: "document:5"},
			},
			expectedWrites: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Mixed writes and deletes",
			index:             0,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
			expectedWrites: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{},
		},
		{
			name:              "Mixed writes and deletes",
			index:             1,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
			expectedWrites: []client.ClientTupleKey{},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
		},
		{
			name:              "Mixed writes and deletes",
			index:             0,
			maxTuplesPerWrite: 5,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
			expectedWrites: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
			},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:3", Relation: "access", Object: "document:3"},
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
		},
		{
			name:              "Mixed writes and deletes",
			index:             1,
			maxTuplesPerWrite: 2,
			writes: []client.ClientTupleKey{
				{User: "user:1", Relation: "access", Object: "document:1"},
				{User: "user:2", Relation: "access", Object: "document:2"},
				{User: "user:3", Relation: "access", Object: "document:3"},
			},
			deletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:4", Relation: "access", Object: "document:4"},
				{User: "user:5", Relation: "access", Object: "document:5"},
				{User: "user:6", Relation: "access", Object: "document:6"},
			},
			expectedWrites: []client.ClientTupleKey{
				{User: "user:3", Relation: "access", Object: "document:3"},
			},
			expectedDeletes: []client.ClientTupleKeyWithoutCondition{
				{User: "user:4", Relation: "access", Object: "document:4"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			writeChunk, deleteChunk := getImportChunk(test.index, test.maxTuplesPerWrite, test.writes, test.deletes)
			assert.Equal(t, test.expectedWrites, writeChunk)
			assert.Equal(t, test.expectedDeletes, deleteChunk)
		})
	}
}
