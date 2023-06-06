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

	"github.com/openfga/fga-cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// MaxStoresLength Limit the stores so that we are not paginating indefinitely.
var MaxStoresLength = 1_000

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stores",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		stores := []openfga.Store{}
		var continuationToken *string
		for {
			options := client.ClientListStoresOptions{
				ContinuationToken: continuationToken,
				PageSize:          nil,
			}
			response, err := fgaClient.ListStores(context.Background()).Options(options).Execute()
			if err != nil {
				fmt.Printf("Failed to list stores due to %v", err)
				os.Exit(1)
			}

			stores = append(stores, *response.Stores...)

			if continuationToken == nil || len(stores) > MaxStoresLength {
				break
			}

			continuationToken = response.ContinuationToken
		}

		storesJSON, err := json.Marshal(stores)
		if err != nil {
			fmt.Printf("Failed to list stores due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(storesJSON))
	},
}

func init() {
}
