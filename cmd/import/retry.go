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
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// statusCmd represents the import command.
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry a import job",
	Long:  "Retry a import job",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		conn, err := storage.NewDatabase()
		bulkJobID, err := cmd.Flags().GetString("job-id")
		if err != nil {
			return fmt.Errorf("failed to get job-id: %w", err)
		}

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
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

		err = job.ImportTuples(conn, bulkJobID, fgaClient, initialRequestRate, maxRequests, rampInterval)
		if err != nil {
			return err
		}

		success, failed, err := storage.GetSummary(conn, bulkJobID)
		fmt.Printf("The status for Job ID - %s: Success - %d, Failed - %d", bulkJobID, success, failed)
		return nil
	},
}

func init() {
	retryCmd.Flags().String("job-id", "", "Job ID")
}
