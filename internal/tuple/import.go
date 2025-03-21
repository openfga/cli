package tuple

import (
	"context"
	"fmt"
	"strings"
	"time"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/rung/go-safecast"

	"github.com/openfga/cli/internal/requests"
)

const (
	// MaxTuplesPerWrite Limit the tuples in a single batch.
	MaxTuplesPerWrite = 1

	// MaxParallelRequests Limit the parallel writes to the API.
	MaxParallelRequests = 10

	// DefaultMinRPS Default minimum requests per second.
	DefaultMinRPS = 1
)

type failedWriteResponse struct {
	TupleKey client.ClientTupleKey `json:"tuple_key"`
	Reason   string                `json:"reason"`
}

type ImportResponse struct {
	Successful []client.ClientTupleKey `json:"successful"`
	Failed     []failedWriteResponse   `json:"failed"`
}

// ImportTuples receives a client.ClientWriteRequest and imports the tuples to the store. It can be used to import
// either writes or deletes.
// It returns a pointer to an ImportResponse and an error.
// The ImportResponse contains the tuples that were successfully imported and the tuples that failed to be imported.
// Deletes and writes are put together in the same ImportResponse.
func ImportTuples(ctx context.Context, fgaClient client.SdkClient,
	minRPS, maxRPS, rampUpPeriodInSec, maxTuplesPerWrite, maxParallelRequests int,
	body client.ClientWriteRequest,
) (*ImportResponse, error) {
	maxTuplesPerWrite32, err := safecast.Int32(maxTuplesPerWrite)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maxTuplesPerWrite due to %w", err)
	}

	maxParallelRequests32, err := safecast.Int32(maxParallelRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maxParallelRequests due to %w", err)
	}

	options := client.ClientWriteOptions{
		Transaction: &client.TransactionOptions{
			Disable:             false,
			MaxPerChunk:         maxTuplesPerWrite32,
			MaxParallelRequests: maxParallelRequests32,
		},
	}

	// If RPS values are 0, then fallback to the previous way of importing
	if minRPS == 0 || maxRPS == 0 {
		return importTuplesWithoutRPS(ctx, fgaClient, body, options)
	}

	return importTuplesWithRPS(ctx, fgaClient,
		minRPS, maxRPS, rampUpPeriodInSec, maxTuplesPerWrite, maxParallelRequests,
		body, options)
}

func importTuplesWithoutRPS(
	ctx context.Context, fgaClient client.SdkClient, body client.ClientWriteRequest, options client.ClientWriteOptions,
) (*ImportResponse, error) {
	response, err := fgaClient.Write(ctx).Body(body).Options(options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to import tuples due to %w", err)
	}

	successful, failed := processWritesAndDeletes(response)
	result := ImportResponse{
		Successful: successful,
		Failed:     failed,
	}

	return &result, nil
}

// importTuplesWithRPS imports tuples to the store with rate limiting.
// It receives a context, an FGA client, rate limiting parameters, and a write request body.
// It returns a pointer to an ImportResponse and an error.
//
// Parameters:
// - ctx: context.Context - The context for the request.
// - fgaClient: client.SdkClient - The FGA client to use for the request.
// - minRPS: int - The minimum requests per second.
// - maxRPS: int - The maximum requests per second.
// - rampUpPeriodInSec: int - The ramp-up period in seconds.
// - maxTuplesPerWrite: int - The maximum number of tuples per write request.
// - maxParallelRequests: int - The maximum number of parallel requests.
// - body: client.ClientWriteRequest - The write request body containing tuples to write or delete.
// - options: client.ClientWriteOptions - The options for the write request.
//
// Returns:
// - *ImportResponse: A pointer to the ImportResponse containing successful and failed tuples.
// - error: An error if the import fails.
func importTuplesWithRPS(ctx context.Context, fgaClient client.SdkClient,
	minRPS, maxRPS, rampUpPeriodInSec, maxTuplesPerWrite, maxParallelRequests int,
	body client.ClientWriteRequest, options client.ClientWriteOptions,
) (*ImportResponse, error) {
	result := ImportResponse{}
	writes := body.Writes
	deletes := body.Deletes
	numRequests := (len(writes) + len(deletes)) / maxTuplesPerWrite

	fmt.Printf(
		"Importing tuples: writing %d tuples and deleting %d tuples over %v requests\n",
		len(writes),
		len(deletes),
		numRequests,
	)

	reqs := make([]func() error, numRequests)

	for requestIndex := range numRequests {
		writeChunk, deleteChunk := getImportChunk(requestIndex, maxTuplesPerWrite, writes, deletes)
		if len(writeChunk)+len(deleteChunk) == 0 {
			fmt.Printf("Failed to import tuples due to empty write chunk index %v\n", requestIndex)

			reqs[requestIndex] = func() error { return nil }

			break
		}

		reqs[requestIndex] = func() error {
			request := fgaClient.Write(ctx).Body(client.ClientWriteRequest{
				Writes:  writeChunk,
				Deletes: deleteChunk,
			}).Options(options)

			response, err := request.Execute()
			if err != nil {
				fmt.Printf("Error writing chunk: %v\n", err)

				return err //nolint:wrapcheck
			}

			successfulWrites, failedWrites := processWrites(response.Writes)
			successfulDeletes, failedDeletes := processDeletes(response.Deletes)

			result.Successful = append(result.Successful, successfulWrites...)
			result.Successful = append(result.Successful, successfulDeletes...)
			result.Failed = append(result.Failed, failedWrites...)
			result.Failed = append(result.Failed, failedDeletes...)

			return nil
		}
	}

	if err := requests.RampUpAPIRequests(
		ctx, minRPS, maxRPS, rampUpPeriodInSec, time.Second, maxParallelRequests, reqs,
	); err != nil {
		return nil, fmt.Errorf("failed to import tuples due to %w", err)
	}

	return &result, nil
}

// getImportChunk returns a chunk of tuples to write.
// It receives an index, the maximum number of tuples per write, and the writes and deletes to import,
// and based on that returns the chunk of tuples to write/delete.
// It does that by filling the buckets with the writes first and then when out of writes, fills the rest with deletes.
func getImportChunk(
	index, maxTuplesPerWrite int,
	writes []client.ClientTupleKey, deletes []client.ClientTupleKeyWithoutCondition) (
	[]client.ClientTupleKey, []client.ClientTupleKeyWithoutCondition,
) {
	start := index * maxTuplesPerWrite
	end := start + maxTuplesPerWrite

	writeChunk := []client.ClientTupleKey{}
	deleteChunk := []client.ClientTupleKeyWithoutCondition{}

	if start < len(writes) {
		if end > len(writes) {
			end = len(writes)
		}

		writeChunk = writes[start:end]
	}

	if len(deletes) == 0 || len(writeChunk) == maxTuplesPerWrite {
		return writeChunk, deleteChunk
	}

	indexOffset := index - len(writes)/maxTuplesPerWrite
	extraWrites := len(writes) % maxTuplesPerWrite
	start = indexOffset * maxTuplesPerWrite
	end = start + maxTuplesPerWrite - extraWrites

	if start < len(deletes) {
		if end > len(deletes) {
			end = len(deletes)
		}

		deleteChunk = deletes[start:end]
	}

	return writeChunk, deleteChunk
}

func extractErrMsg(err error) string {
	errorMsg := err.Error()
	startIndex := strings.Index(errorMsg, "error message:")

	if startIndex == -1 {
		return errorMsg
	}

	errorMsg = errorMsg[startIndex:]
	errorMsg = strings.TrimSpace(errorMsg)

	return errorMsg
}

func processWritesAndDeletes(
	response *client.ClientWriteResponse,
) ([]client.ClientTupleKey, []failedWriteResponse) {
	successfulWrites, failedWrites := processWrites(response.Writes)
	successfulDeletes, failedDeletes := processDeletes(response.Deletes)

	return append(successfulWrites, successfulDeletes...), append(failedWrites, failedDeletes...)
}

func processWrites(
	writes []client.ClientWriteRequestWriteResponse,
) ([]client.ClientTupleKey, []failedWriteResponse) {
	var (
		successfulWrites []client.ClientTupleKey
		failedWrites     []failedWriteResponse
	)

	for _, write := range writes {
		if write.Status == client.SUCCESS {
			successfulWrites = append(successfulWrites, write.TupleKey)
		} else {
			reason := extractErrMsg(write.Error)
			failedWrites = append(failedWrites, failedWriteResponse{
				TupleKey: write.TupleKey,
				Reason:   reason,
			})
		}
	}

	return successfulWrites, failedWrites
}

func processDeletes(
	deletes []client.ClientWriteRequestDeleteResponse,
) ([]client.ClientTupleKey, []failedWriteResponse) {
	var (
		successfulDeletes []client.ClientTupleKey
		failedDeletes     []failedWriteResponse
	)

	for _, del := range deletes {
		deletedTupleKey := openfga.TupleKey{
			Object:   del.TupleKey.Object,
			Relation: del.TupleKey.Relation,
			User:     del.TupleKey.User,
		}

		if del.Status == client.SUCCESS {
			successfulDeletes = append(successfulDeletes, deletedTupleKey)
		} else {
			reason := extractErrMsg(del.Error)
			failedDeletes = append(failedDeletes, failedWriteResponse{
				TupleKey: deletedTupleKey,
				Reason:   reason,
			})
		}
	}

	return successfulDeletes, failedDeletes
}
