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
	"encoding/json"
	"fmt"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"os"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete Relationship Tuples",
	Long:    "Delete relationship tuples from the store.",
	Example: "fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 && cmd.Flags().Changed("file") == false {
			return fmt.Errorf("you need to specify either 3 arguments or a file")
		}
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}
		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to parse file name due to %w", err)
		}
		if fileName != "" {
			var tuples []client.ClientTupleKey

			data, err := os.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("failed to read file %s due to %w", fileName, err)
			}

			err = json.Unmarshal(data, &tuples)
			if err != nil {
				return fmt.Errorf("failed to parse input tuples due to %w", err)
			}
			err = deleteTuples(fgaClient, tuples)
			if err != nil {
				return err
			}
			return output.Display(output.EmptyStruct{}) //nolint:wrapcheck
		}
		body := &client.ClientDeleteTuplesBody{
			client.ClientTupleKey{
				User:     args[0],
				Relation: args[1],
				Object:   args[2],
			},
		}
		options := &client.ClientWriteOptions{}
		_, err = fgaClient.DeleteTuples(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			return fmt.Errorf("failed to delete tuples due to %w", err)
		}

		return output.Display(output.EmptyStruct{}) //nolint:wrapcheck
	},
}

func deleteTuples(fgaClient *client.OpenFgaClient, tuples []client.ClientTupleKey) error {
	_, err := fgaClient.DeleteTuples(context.Background()).Body(tuples).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete tuples due to %w", err)
	}
	return nil
}

func init() {
	deleteCmd.Flags().String("file", "", "Tuples file")
}
