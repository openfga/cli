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
	"fmt"
	"os"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Create Relationship Tuples",
	Long:  "Add relationship tuples to the store.",
	Args:  ExactArgsOrFlag(3, "file"), //nolint:gomnd
	Example: `fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap
fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json
fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.yaml
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			maxTuplesPerWrite, err := cmd.Flags().GetInt("max-tuples-per-write")
			if err != nil {
				return fmt.Errorf("failed to parse max tuples per write due to %w", err)
			}

			maxParallelRequests, err := cmd.Flags().GetInt("max-parallel-requests")
			if err != nil {
				return fmt.Errorf("failed to parse parallel requests due to %w", err)
			}

			var tuples []client.ClientTupleKey

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
			response, err := ImportTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests)
			if err != nil {
				return err
			}

			return output.Display(*response) //nolint:wrapcheck
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

func init() {
	writeCmd.Flags().String("model-id", "", "Model ID")
	writeCmd.Flags().String("file", "", "Tuples file")
	writeCmd.Flags().Int("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	writeCmd.Flags().Int("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll
}
