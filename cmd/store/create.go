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
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func create(fgaClient client.SdkClient, storeName string) (string, error) {
	body := client.ClientCreateStoreRequest{Name: storeName}

	store, err := fgaClient.CreateStore(context.Background()).Body(body).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create store %v due to %w", storeName, err)
	}

	storeJSON, err := json.Marshal(store)
	if err != nil {
		return "", fmt.Errorf("store %v created, but failed to be printed due to %w", *store.Id, err)
	}

	return string(storeJSON), nil
}

// createCmd represents the store create command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and initialize a store.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}
		storeName, _ := cmd.Flags().GetString("name")
		output, err := create(fgaClient, storeName)
		if err != nil {
			return err
		}
		fmt.Print(output)

		return nil
	},
}

func init() {
	createCmd.Flags().String("name", "", "Store Name")
	err := createCmd.MarkFlagRequired("name")
	if err != nil { //nolint:wsl
		fmt.Print(err)
		os.Exit(1)
	}
}
