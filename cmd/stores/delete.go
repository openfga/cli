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
	"os"

	"github.com/openfga/fga-cli/lib/cmd-utils"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete store",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		_, err = fgaClient.DeleteStore(context.Background()).Execute()
		if err != nil {
			fmt.Printf("Failed to create store %v due to %v", clientConfig.StoreID, err)
			os.Exit(1)
		}

		fmt.Printf(`{}`)
	},
}

func init() {
	deleteCmd.Flags().String("store-id", "", "Store ID")
	err := deleteCmd.MarkFlagRequired("store-id")
	if err != nil { //nolint:wsl
		fmt.Print(err)
		os.Exit(1)
	}
}
