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
package stores

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// MaxStoresPagesLength Limit the pages of stores so that we are not paginating indefinitely.
var MaxStoresPagesLength = 20 // up to 1000 records

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stores",
	Long:  `Get a list of stores.`,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		maxPages, _ := cmd.Flags().GetInt("max-pages")
		if err != nil {
			fmt.Printf("Failed to list models due to %v", err)
			os.Exit(1)
		}
		stores := []openfga.Store{}
		continuationToken := ""
		pageIndex := 0
		for {
			options := client.ClientListStoresOptions{
				ContinuationToken: &continuationToken,
			}
			response, err := fgaClient.ListStores(context.Background()).Options(options).Execute()
			if err != nil {
				fmt.Printf("Failed to list stores due to %v", err)
				os.Exit(1)
			}

			stores = append(stores, *response.Stores...)
			pageIndex++
			if response.ContinuationToken == nil || *response.ContinuationToken == continuationToken || pageIndex >= maxPages {
				break
			}

			continuationToken = *response.ContinuationToken
		}

		storesJSON, err := json.Marshal(openfga.ListStoresResponse{Stores: &stores})
		if err != nil {
			fmt.Printf("Failed to list stores due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(storesJSON))
	},
}

func init() {
	listCmd.Flags().Int("max-pages", MaxStoresPagesLength, "Max number of pages to get.")
}
