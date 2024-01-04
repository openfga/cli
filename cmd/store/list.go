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

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// MaxStoresPagesLength Limit the pages of stores so that we are not paginating indefinitely.
var MaxStoresPagesLength = 20 // up to 1000 records

func listStores(fgaClient client.SdkClient, maxPages int) (*openfga.ListStoresResponse, error) {
	stores := []openfga.Store{}
	continuationToken := ""
	pageIndex := 0

	for {
		options := client.ClientListStoresOptions{
			ContinuationToken: &continuationToken,
		}

		response, err := fgaClient.ListStores(context.Background()).Options(options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list stores due to %w", err)
		}

		stores = append(stores, response.Stores...)
		pageIndex++

		if response.ContinuationToken == "" || pageIndex >= maxPages {
			break
		}

		continuationToken = response.ContinuationToken
	}

	return &openfga.ListStoresResponse{Stores: stores}, nil
}

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List Stores",
	Long:    `Get a list of stores.`,
	Example: "fga store list",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		maxPages, _ := cmd.Flags().GetInt("max-pages")
		if err != nil {
			return fmt.Errorf("failed to parse max pages due to %w", err)
		}

		response, err := listStores(fgaClient, maxPages)
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	listCmd.Flags().Int("max-pages", MaxStoresPagesLength, "Max number of pages to get.")
}
