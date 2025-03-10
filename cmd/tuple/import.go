/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tuple

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

const (
	// MaxTuplesPerWrite Limit the tuples in a single batch.
	MaxTuplesPerWrite = 1

	// MaxParallelRequests Limit the parallel writes to the API.
	MaxParallelRequests = 10
)

type failedWriteResponse struct {
	TupleKey interface{} `json:"tuple_key"`
	Reason   string      `json:"reason"`
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
func ImportTuples(
	fgaClient client.SdkClient,
	body client.ClientWriteRequest,
	maxTuplesPerWrite int32,
	maxParallelRequests int32,
	minRPS int32,
	maxRPS int32,
	rampupPeriodInSec int32,
) (*ImportResponse, error) {
	options := client.ClientWriteOptions{
		Transaction: &client.TransactionOptions{
			Disable:             true,
			MaxPerChunk:         maxTuplesPerWrite,
			MaxParallelRequests: maxParallelRequests,
		},
	}

	// If RPS control is not enabled, use the standard write method
	if minRPS <= 0 || maxRPS <= 0 || rampupPeriodInSec <= 0 {
		response, err := fgaClient.Write(context.Background()).Body(body).Options(options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to import tuples due to %w", err)
		}

		successfulWrites, failedWrites := processWrites(response.Writes)
		successfulDeletes, failedDeletes := processDeletes(response.Deletes)

		result := ImportResponse{
			Successful: append(successfulWrites, successfulDeletes...),
			Failed:     append(failedWrites, failedDeletes...),
		}

		return &result, nil
	}

	// RPS control is enabled, implement rate-limited writes with ramp-up
	return importTuplesWithRateLimit(fgaClient, body, options, minRPS, maxRPS, rampupPeriodInSec)
}

// importTuplesWithRateLimit imports tuples with rate limiting and ramp-up
func importTuplesWithRateLimit(
	fgaClient client.SdkClient,
	body client.ClientWriteRequest,
	options client.ClientWriteOptions,
	minRPS int32,
	maxRPS int32,
	rampupPeriodInSec int32,
) (*ImportResponse, error) {
	ctx := context.Background()

	// Prepare result containers
	var successfulWrites []client.ClientTupleKey
	var failedWrites []failedWriteResponse
	var successfulDeletes []client.ClientTupleKey
	var failedDeletes []failedWriteResponse

	// Calculate total number of tuples to write
	totalTuples := len(body.Writes) + len(body.Deletes)
	if totalTuples == 0 {
		return &ImportResponse{
			Successful: []client.ClientTupleKey{},
			Failed:     []failedWriteResponse{},
		}, nil
	}

	// Create batches of tuples to write
	// Each batch will contain a single tuple for now
	// This could be optimized in the future to use maxTuplesPerWrite
	var writeBatches []client.ClientWriteRequest

	// Add write tuples to batches
	for _, tuple := range body.Writes {
		// Convert to ClientTupleKeyWithoutCondition if needed
		tupleWithoutCondition, ok := tuple.(client.ClientTupleKeyWithoutCondition)
		if ok {
			writeBatches = append(writeBatches, client.ClientWriteRequest{
				Writes: []client.ClientTupleKeyWithoutCondition{tupleWithoutCondition},
			})
		} else {
			// Assume it's already a ClientTupleKey
			writeBatches = append(writeBatches, client.ClientWriteRequest{
				Writes: []client.ClientTupleKey{tuple},
			})
		}
	}

	// Add delete tuples to batches
	for _, tuple := range body.Deletes {
		// Convert to ClientTupleKeyWithoutCondition if needed
		tupleWithoutCondition, ok := tuple.(client.ClientTupleKeyWithoutCondition)
		if ok {
			writeBatches = append(writeBatches, client.ClientWriteRequest{
				Deletes: []client.ClientTupleKeyWithoutCondition{tupleWithoutCondition},
			})
		} else {
			// Assume it's already a ClientTupleKey
			writeBatches = append(writeBatches, client.ClientWriteRequest{
				Deletes: []client.ClientTupleKey{tuple},
			})
		}
	}

	// Calculate ramp-up parameters
	rampupPeriod := time.Duration(rampupPeriodInSec) * time.Second
	startTime := time.Now()
	endRampupTime := startTime.Add(rampupPeriod)

	// Process each batch with rate limiting
	for i, batch := range writeBatches {
		// Calculate current RPS based on elapsed time and ramp-up period
		currentTime := time.Now()
		if currentTime.After(endRampupTime) {
			// Ramp-up period has passed, use max RPS
			time.Sleep(time.Second / time.Duration(maxRPS))
		} else {
			// Still in ramp-up period, calculate current RPS
			elapsedRatio := float64(currentTime.Sub(startTime)) / float64(rampupPeriod)
			currentRPS := minRPS + int32(float64(maxRPS-minRPS)*elapsedRatio)
			if currentRPS < minRPS {
				currentRPS = minRPS
			}
			if currentRPS > maxRPS {
				currentRPS = maxRPS
			}

			// Sleep to maintain the current RPS
			time.Sleep(time.Second / time.Duration(currentRPS))
		}

		// Execute the write request
		response, err := fgaClient.Write(ctx).Body(batch).Options(options).Execute()
		if err != nil {
			// If the entire request failed, add all tuples in this batch to failed
			if len(batch.Writes) > 0 {
				for _, tuple := range batch.Writes {
					failedWrites = append(failedWrites, failedWriteResponse{
						TupleKey: tuple,
						Reason:   extractErrMssg(err),
					})
				}
			}
			if len(batch.Deletes) > 0 {
				for _, tuple := range batch.Deletes {
					failedDeletes = append(failedDeletes, failedWriteResponse{
						TupleKey: tuple,
						Reason:   extractErrMssg(err),
					})
				}
			}
			continue
		}

		// Process successful and failed writes/deletes
		sw, fw := processWrites(response.Writes)
		successfulWrites = append(successfulWrites, sw...)
		failedWrites = append(failedWrites, fw...)

		sd, fd := processDeletes(response.Deletes)
		successfulDeletes = append(successfulDeletes, sd...)
		failedDeletes = append(failedDeletes, fd...)

		// Log progress (optional)
		if i > 0 && i%100 == 0 {
			fmt.Fprintf(os.Stderr, "Processed %d/%d tuples\n", i, totalTuples)
		}
	}

	result := ImportResponse{
		Successful: append(successfulWrites, successfulDeletes...),
		Failed:     append(failedWrites, failedDeletes...),
	}

	return &result, nil
}

func extractErrMssg(err error) string {
	errorMsg := err.Error()
	startIndex := strings.Index(errorMsg, "error message:")

	if startIndex == -1 {
		return errorMsg
	}

	errorMsg = errorMsg[startIndex:]
	errorMsg = strings.TrimSpace(errorMsg)

	return errorMsg
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
			reason := extractErrMssg(write.Error)
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

	for _, delete := range deletes {
		deletedTupleKey := openfga.TupleKey{
			Object:   delete.TupleKey.Object,
			Relation: delete.TupleKey.Relation,
			User:     delete.TupleKey.User,
		}

		if delete.Status == client.SUCCESS {
			successfulDeletes = append(successfulDeletes, deletedTupleKey)
		} else {
			reason := extractErrMssg(delete.Error)
			failedDeletes = append(failedDeletes, failedWriteResponse{
				TupleKey: deletedTupleKey,
				Reason:   reason,
			})
		}
	}

	return successfulDeletes, failedDeletes
}

// importCmd represents the import command.
var importCmd = &cobra.Command{
	Use:        "import",
	Short:      "Import Relationship Tuples",
	Deprecated: "use the write/delete command with the flag --file instead",
	Long: "Imports Relationship Tuples to the store. " +
		"This will write the tuples in chunks and at the end will report the tuple chunks that failed.\n\n" +
		"Rate Limiting:\n" +
		"You can control the rate at which tuples are written using the following flags:\n" +
		"- min-rps: Minimum requests per second for writes\n" +
		"- max-rps: Maximum requests per second for writes\n" +
		"- rampup-period-in-sec: Period in seconds to ramp up from min-rps to max-rps\n\n" +
		"If any of these flags are provided, all three must be provided with positive values. " +
		"The command will start writing tuples at the min-rps rate and gradually increase to " +
		"max-rps over the specified rampup period. If all tuples are written before the rampup " +
		"period ends, the command will exit. If the rampup period ends and there are still tuples " +
		"to write, the command will continue writing at the max-rps rate until all tuples are written.",
	Example: `  fga tuple import --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.yaml
  fga tuple import --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.yaml --min-rps=10 --max-rps=50 --rampup-period-in-sec=60`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to parse file name due to %w", err)
		}

		maxTuplesPerWrite, err := cmd.Flags().GetInt32("max-tuples-per-write")
		if err != nil {
			return fmt.Errorf("failed to parse max tuples per write due to %w", err)
		}

		maxParallelRequests, err := cmd.Flags().GetInt32("max-parallel-requests")
		if err != nil {
			return fmt.Errorf("failed to parse parallel requests due to %w", err)
		}

		// Extract RPS control parameters
		minRPS, err := cmd.Flags().GetInt32("min-rps")
		if err != nil {
			return fmt.Errorf("failed to parse min-rps: %w", err)
		}

		maxRPS, err := cmd.Flags().GetInt32("max-rps")
		if err != nil {
			return fmt.Errorf("failed to parse max-rps: %w", err)
		}

		rampupPeriod, err := cmd.Flags().GetInt32("rampup-period-in-sec")
		if err != nil {
			return fmt.Errorf("failed to parse rampup-period-in-sec: %w", err)
		}

		// Validate RPS parameters - if one is provided, all three should be required
		if minRPS > 0 || maxRPS > 0 || rampupPeriod > 0 {
			if minRPS <= 0 || maxRPS <= 0 || rampupPeriod <= 0 {
				return errors.New("if any of min-rps, max-rps, or rampup-period-in-sec is provided, all three must be provided with positive values") //nolint:goerr113
			}

			if minRPS > maxRPS {
				return errors.New("min-rps cannot be greater than max-rps") //nolint:goerr113
			}
		}

		tuples := []client.ClientTupleKey{}

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read file %s due to %w", fileName, err)
		}

		err = yaml.Unmarshal(data, &tuples)
		if err != nil {
			return fmt.Errorf("failed to parse input tuples due to %w", err)
		}

		writeRequest := client.ClientWriteRequest{
			Writes: tuples,
		}

		result, err := ImportTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests, minRPS, maxRPS, rampupPeriod)
		if err != nil {
			return err
		}

		return output.Display(*result)
	},
}

func init() {
	importCmd.Flags().String("model-id", "", "Model ID")
	importCmd.Flags().String("file", "", "Tuples file")
	importCmd.Flags().Int32("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	importCmd.Flags().Int32("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll
	importCmd.Flags().Int32("min-rps", 0, "Minimum requests per second for writes")
	importCmd.Flags().Int32("max-rps", 0, "Maximum requests per second for writes")
	importCmd.Flags().Int32("rampup-period-in-sec", 0, "Period in seconds to ramp up from min-rps to max-rps")
}
