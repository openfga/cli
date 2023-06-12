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
package models

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

// MaxModelsPagesLength Limit the models so that we are not paginating indefinitely.
var MaxModelsPagesLength = 20

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Read Authorization Models",
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		maxPages, err := cmd.Flags().GetInt("max-pages")
		if err != nil {
			fmt.Printf("Failed to list models due to %v", err)
			os.Exit(1)
		}
		models := []openfga.AuthorizationModel{}
		var continuationToken *string
		pageIndex := 0
		for {
			options := client.ClientReadAuthorizationModelsOptions{
				ContinuationToken: continuationToken,
			}
			response, err := fgaClient.ReadAuthorizationModels(context.Background()).Options(options).Execute()
			if err != nil {
				fmt.Printf("Failed to list models due to %v", err)
				os.Exit(1)
			}

			models = append(models, *response.AuthorizationModels...)

			pageIndex++
			if continuationToken == nil || pageIndex > maxPages {
				break
			}

			continuationToken = response.ContinuationToken
		}

		modelsJSON, err := json.Marshal(openfga.ReadAuthorizationModelsResponse{AuthorizationModels: &models})
		if err != nil {
			fmt.Printf("Failed to list models due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(modelsJSON))
	},
}

func init() {
	listCmd.Flags().Int("max-pages", MaxModelsPagesLength, "Max number of pages to get.")
}
