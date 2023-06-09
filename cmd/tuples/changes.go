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
package tuples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/fga-cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// MaxReadChangesPagesLength Limit the changes so that we are not paginating indefinitely.
var MaxReadChangesPagesLength = 20

// changesCmd represents the changes command.
var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Read Relationship Tuple Changes (Watch)",
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		maxPages, err := cmd.Flags().GetInt("max-pages")
		if err != nil {
			fmt.Printf("Failed to get tuple changes due to %v", err)
			os.Exit(1)
		}
		selectedType, err := cmd.Flags().GetString("type")
		if err != nil {
			fmt.Printf("Failed to get tuple changes due to %v", err)
			os.Exit(1)
		}
		changes := []openfga.TupleChange{}
		var continuationToken *string
		pageIndex := 0
		for {
			body := &client.ClientReadChangesRequest{
				Type: selectedType,
			}
			options := &client.ClientReadChangesOptions{}
			response, err := fgaClient.ReadChanges(context.Background()).Body(*body).Options(*options).Execute()
			if err != nil {
				fmt.Printf("Failed to get tuple changes due to %v", err)
				os.Exit(1)
			}

			changes = append(changes, *response.Changes...)
			pageIndex++
			if continuationToken == nil || pageIndex >= maxPages {
				break
			}

			continuationToken = response.ContinuationToken
		}

		changesJSON, err := json.Marshal(openfga.ReadChangesResponse{Changes: &changes})
		if err != nil {
			fmt.Printf("Failed to tuple changes due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(changesJSON))
	},
}

func init() {
	changesCmd.Flags().String("type", "", "Type to restrict the changes by.")
	changesCmd.Flags().Int("max-pages", MaxReadChangesPagesLength, "Max number of pages to get.")
}
