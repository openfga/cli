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

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/cli/lib/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:     "write",
	Short:   "Create Relationship Tuples",
	Args:    cobra.ExactArgs(3), //nolint:gomnd
	Example: "fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
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
}
