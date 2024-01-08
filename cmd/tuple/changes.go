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

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// MaxReadChangesPagesLength Limit the changes so that we are not paginating indefinitely.
var MaxReadChangesPagesLength = 20

func readChanges(
	fgaClient client.SdkClient, maxPages int, selectedType string, continuationToken string,
) (*openfga.ReadChangesResponse, error) {
	changes := []openfga.TupleChange{}
	pageIndex := 0

	for {
		body := &client.ClientReadChangesRequest{
			Type: selectedType,
		}
		options := &client.ClientReadChangesOptions{
			ContinuationToken: &continuationToken,
		}

		response, err := fgaClient.ReadChanges(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get tuple changes due to %w", err)
		}

		changes = append(changes, response.Changes...)
		previousContinuationToken := continuationToken
		continuationToken = *response.ContinuationToken
		pageIndex++

		if response.ContinuationToken == nil ||
			continuationToken == previousContinuationToken ||
			pageIndex >= maxPages {
			break
		}
	}

	return &openfga.ReadChangesResponse{Changes: changes, ContinuationToken: &continuationToken}, nil
}

// changesCmd represents the changes command.
var changesCmd = &cobra.Command{
	Use:     "changes",
	Short:   "Read Relationship Tuple Changes (Watch)",
	Long:    "Get a list of relationship tuple changes (Writes and Deletes) across time.",
	Example: "fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --type document --continuation-token=MXw=",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		maxPages, err := cmd.Flags().GetInt("max-pages")
		if err != nil {
			return fmt.Errorf("failed to parse max pages due to %w", err)
		}

		selectedType, err := cmd.Flags().GetString("type")
		if err != nil {
			return fmt.Errorf("failed to get tuple changes due to %w", err)
		}

		continuationToken, err := cmd.Flags().GetString("continuation-token")
		if err != nil {
			return fmt.Errorf("failed to get tuple changes due to %w", err)
		}

		response, err := readChanges(fgaClient, maxPages, selectedType, continuationToken)
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	changesCmd.Flags().String("type", "", "Type to restrict the changes by.")
	changesCmd.Flags().Int("max-pages", MaxReadChangesPagesLength, "Max number of pages to get.")
	changesCmd.Flags().String("continuation-token", "", "Continuation token to start changes from.")
}
