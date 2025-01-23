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
	"fmt"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/job"
	"github.com/openfga/cli/internal/storage"
	"github.com/openfga/cli/internal/tuplefile"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// ImportJobTuples receives a client.ClientWriteRequest and imports the tuples to the store. It can be used to import
// either writes or deletes.
// It returns a pointer to an ImportResponse and an error.
// The ImportResponse contains the tuples that were successfully imported and the tuples that failed to be imported.
// Deletes and writes are put together in the same ImportResponse.
func ImportJobTuples(
	fgaClient client.SdkClient,
	tuples []client.ClientTupleKey,
	storeID string,
	requestRate int,
	maxRequests int,
	rampIntervalInSeconds int64,
) error {
	conn, err := storage.NewDatabase()
	if err != nil {
		return err
	}
	bulkJobID, err := job.CreateJob(conn, storeID, tuples)
	if err != nil {
		return err
	}
	fmt.Printf("Job created successfully - %s\n", bulkJobID)

	err = job.ImportTuples(conn, bulkJobID, fgaClient, requestRate, maxRequests, rampIntervalInSeconds)
	if err != nil {
		return err
	}

	success, failed, err := storage.GetSummary(conn, bulkJobID)
	fmt.Printf("The status for Job ID - %s: Success - %d, Failed - %d", bulkJobID, success, failed)
	return nil
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

		initialRequestRate, err := cmd.Flags().GetInt("initial-request-rate")
		if initialRequestRate <= 0 || err != nil {
			initialRequestRate = 20
		}

		maxRequests, err := cmd.Flags().GetInt("max-requests")
		if maxRequests <= 0 || err != nil {
			maxRequests = 2000
		}

		rampInterval, err := cmd.Flags().GetInt64("ramp-interval-seconds")
		if rampInterval <= 0 || err != nil {
			rampInterval = 120
		}

		var tuples []client.ClientTupleKey
		tuples, err = tuplefile.ReadTupleFile(fileName)
		if err != nil {
			return err //nolint:wrapcheck
		}

		err = ImportJobTuples(fgaClient, tuples, storeID, initialRequestRate, maxRequests, rampInterval)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	createCmd.Flags().String("file", "", "Tuples file")
	createCmd.Flags().String("store-id", "", "Store ID")
}
