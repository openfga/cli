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

package _import

import (
	"context"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storage"
	"github.com/openfga/cli/internal/tuplefile"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
	"sync/atomic"
)

type CreateImportJobResponse struct {
	JobId string `json:"job_id"`
}

// ImportJobTuples receives a client.ClientWriteRequest and imports the tuples to the store. It can be used to import
// either writes or deletes.
// It returns a pointer to an ImportResponse and an error.
// The ImportResponse contains the tuples that were successfully imported and the tuples that failed to be imported.
// Deletes and writes are put together in the same ImportResponse.
func ImportJobTuples(
	fgaClient client.SdkClient,
	tuples []client.ClientTupleKey,
	storeID string,
) (*CreateImportJobResponse, error) {
	bulkJobID := ulid.Make().String()
	db, err := storage.NewDatabase()
	if err != nil {
		return nil, err
	}
	err = storage.InsertTuples(db, bulkJobID, storeID, tuples)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Job created successfully - %s\n", bulkJobID)

	notInsertedTuplesCount, insertedTuplesCount, err := storage.GetTotalAndRemainingTuples(db, bulkJobID)
	totalTuplesCount := insertedTuplesCount + notInsertedTuplesCount
	if err != nil {
		return nil, err
	}
	completedTuples := atomic.Int64{}
	completedTuples.Store(insertedTuplesCount)

	for i := completedTuples; i.Load() < totalTuplesCount; {
		remainingTuples, err := storage.GetRemainingTuples(db, bulkJobID, 3)
		if err != nil {
			return nil, err
		}
		for _, tuple := range remainingTuples {
			go func() {
				_, err = fgaClient.
					WriteTuples(context.Background()).
					Body(client.ClientWriteTuplesBody{tuple.Tuple}).
					Options(client.ClientWriteOptions{}).
					Execute()
				if err != nil {
					err.Error()
				} else {
					println("Success")
				}
				completedTuples.Add(1)
			}()
		}
	}

	// Write 1 tuple per request
	// Start with 20 requests per second
	// Slowly ramp up - find ramp up logic
	// Each time a request is successful we have to write it to a file
	// --Resuming
	// If resuming, we can ignore all the successful writes, start with failed writes
	// --Failed
	// Once fully completed, we can show a summary
	// No. of successful writes, number of failed writes
	// --While executing we can show
	// Percentage completed, current RPS, expected time to complete
	return &CreateImportJobResponse{JobId: bulkJobID}, nil
}

// createCmd represents the import command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Import Relationship Tuples as a job",
	Long: "Imports Relationship Tuples to the store. " +
		"This will write the tuples in chunks and at the end will report the tuple chunks that failed.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		storeID, err := cmd.Flags().GetString("store-id")
		if err != nil {
			return fmt.Errorf("failed to get store-id: %w", err)
		}

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to parse file name due to %w", err)
		}

		var tuples []client.ClientTupleKey
		tuples, err = tuplefile.ReadTupleFile(fileName)
		if err != nil {
			return err //nolint:wrapcheck
		}

		result, err := ImportJobTuples(fgaClient, tuples, storeID)
		if err != nil {
			return err
		}

		return output.Display(*result)
	},
}

func init() {
	createCmd.Flags().String("file", "", "Tuples file")
	createCmd.Flags().String("store-id", "", "Store ID")
}
