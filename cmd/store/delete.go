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

package store

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/confirmation"
	"github.com/openfga/cli/internal/output"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete Store",
	Long:    "Mark a store as deleted.",
	Example: "fga store delete --store-id=01H0H015178Y2V4CX10C2KGHF4",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		// First, confirm whether this is intended
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return fmt.Errorf("failed to parse force flag due to %w", err)
		}
		if !force {
			confirmation, err := confirmation.AskForConfirmation("Are you sure you want to delete the store:")
			if err != nil {
				return fmt.Errorf("prompt failed due to %w", err)
			}
			if !confirmation {
				type returnMessage struct {
					Message string
				}

				return output.Display(returnMessage{Message: "Delete store cancelled"}) //nolint:wrapcheck
			}
		}

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}
		_, err = fgaClient.DeleteStore(context.Background()).Execute()
		if err != nil {
			return fmt.Errorf("failed to delete store %v due to %w", clientConfig.StoreID, err)
		}

		return output.Display(output.EmptyStruct{}) //nolint:wrapcheck
	},
}

func init() {
	deleteCmd.Flags().String("store-id", "", "Store ID")
	deleteCmd.Flags().Bool("force", false, "Force delete without confirmation")

	err := deleteCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
