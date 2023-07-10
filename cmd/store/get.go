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
	"os"

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/cli/lib/fga"
	"github.com/openfga/cli/lib/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func getStore(clientConfig fga.ClientConfig, fgaClient client.SdkClient) (*client.ClientGetStoreResponse, error) {
	store, err := fgaClient.GetStore(context.Background()).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get store %v due to %w", clientConfig.StoreID, err)
	}

	return store, nil
}

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get store",
	Long:    ``,
	Example: "fga store get --store-id=01H0H015178Y2V4CX10C2KGHF4",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		response, err := getStore(clientConfig, fgaClient)
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	getCmd.Flags().String("store-id", "", "Store ID")

	err := getCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
