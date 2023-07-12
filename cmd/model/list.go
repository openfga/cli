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
package model

import (
	"context"
	"fmt"
	"os"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// MaxModelsPagesLength Limit the models so that we are not paginating indefinitely.
var MaxModelsPagesLength = 20

func listModels(fgaClient client.SdkClient, maxPages int) (*openfga.ReadAuthorizationModelsResponse, error) {
	// This is needed to ensure empty array is marshaled as [] instead of nil
	models := make([]openfga.AuthorizationModel, 0)

	var continuationToken string

	pageIndex := 0

	for {
		options := client.ClientReadAuthorizationModelsOptions{
			ContinuationToken: &continuationToken,
		}

		response, err := fgaClient.ReadAuthorizationModels(context.Background()).Options(options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list models due to %w", err)
		}

		models = append(models, *response.AuthorizationModels...)

		pageIndex++

		if response.ContinuationToken == nil || *response.ContinuationToken == "" || pageIndex > maxPages {
			break
		}

		continuationToken = *response.ContinuationToken
	}

	return &openfga.ReadAuthorizationModelsResponse{AuthorizationModels: &models}, nil
}

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Read Authorization Models",
	Example: "fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to intialized FGA client due to %w", err)
		}

		maxPages, err := cmd.Flags().GetInt("max-pages")
		if err != nil {
			return fmt.Errorf("failed to parse max pages due to %w", err)
		}

		response, err := listModels(fgaClient, maxPages)
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	listCmd.Flags().Int("max-pages", MaxModelsPagesLength, "Max number of pages to get.")
	listCmd.Flags().String("store-id", "", "Store ID")

	if err := listCmd.MarkFlagRequired("store-id"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/list", err)
		os.Exit(1)
	}
}
