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
	"github.com/openfga/cli/internal/storage"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// statusCmd represents the import command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the import job",
	Long:  "The status command is used to check if the import job is running.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		bulkJobID, err := cmd.Flags().GetString("job-id")
		if err != nil {
			return fmt.Errorf("failed to get job-id: %w", err)
		}
		conn, err := storage.NewDatabase()
		if err != nil {
			return err
		}
		success, failed, err := storage.GetSummary(conn, bulkJobID)
		if err != nil {
			return err
		}
		fmt.Printf("The status for Job ID - %s: Success - %d, Failed - %d", bulkJobID, success, failed)
		return nil
	},
}

func init() {
	statusCmd.Flags().String("job-id", "", "Job ID")
}
