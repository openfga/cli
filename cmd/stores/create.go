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
	. "github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"os"
)

// createCmd represents the store create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and initialize a store.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmd_utils.GetClientConfig(cmd, args)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			panic(err)
		}
		storeName, _ := cmd.Flags().GetString("name")
		body := ClientCreateStoreRequest{Name: storeName}
		store, err := fgaClient.CreateStore(context.Background()).Body(body).Execute()
		if err != nil {
			fmt.Printf("Failed to create store %v due to %v", storeName, err)
			os.Exit(1)
		}

		fmt.Printf(`{"id": "%v", "name":"%v", "created_at": "%v", "updated_at":"%v"}`, *store.Id, *store.Name, store.CreatedAt, store.UpdatedAt)
	},
}

func init() {
	createCmd.Flags().String("name", "", "Store Name")
	createCmd.MarkFlagRequired("name")
}
