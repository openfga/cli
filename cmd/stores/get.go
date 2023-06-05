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
	"fmt"
	cmd_utils "github.com/openfga/fga-cli/lib/cmd-utils"
	"github.com/spf13/cobra"
	"os"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get store",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmd_utils.GetClientConfig(cmd, args)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		store, err := fgaClient.GetStore(context.Background()).Execute()
		if err != nil {
			fmt.Printf("Failed to get store %v due to %v", clientConfig.StoreId, err)
			os.Exit(1)
		}

		fmt.Printf(`{"id": "%v", "name":"%v", "created_at": "%v", "updated_at":"%v"}`, *store.Id, *store.Name, store.CreatedAt, store.UpdatedAt)
	},
}

func init() {
	getCmd.Flags().String("store-id", "", "Store ID")
	getCmd.MarkFlagRequired("store-id")
}
