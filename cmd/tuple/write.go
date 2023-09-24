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
	"os"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:     "write",
	Short:   "Create Relationship Tuples",
	Long:    "Add relationship tuples to the store.",
	Example: "fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap",
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
			tuples := []client.ClientTupleKey{}

			data, err := os.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("failed to read file %s due to %w", fileName, err)
			}

			err = json.Unmarshal(data, &tuples)
			if err != nil {
				return fmt.Errorf("failed to parse input tuples due to %w", err)
			}
			err = writeTuples(fgaClient, tuples)
			if err != nil {
				return err
			}
			return output.Display(output.EmptyStruct{}) //nolint:wrapcheck
		}
		body := &client.ClientWriteTuplesBody{
			client.ClientTupleKey{
				User:     args[0],
				Relation: args[1],
				Object:   args[2],
			},
		}
		options := &client.ClientWriteOptions{}
		_, err = fgaClient.WriteTuples(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			return fmt.Errorf("failed to write tuples due to %w", err)
		}

		return output.Display(output.EmptyStruct{}) //nolint:wrapcheck
	},
}

func writeTuples(fgaClient *client.OpenFgaClient, tuples []client.ClientTupleKey) error {
	_, err := fgaClient.WriteTuples(context.Background()).Body(tuples).Execute()
	if err != nil {
		return fmt.Errorf("failed to write tuples due to %w", err)
	}
	return nil
}

func init() {
	writeCmd.Flags().String("model-id", "", "Model ID")
	writeCmd.Flags().String("file", "", "Tuples file")
}
