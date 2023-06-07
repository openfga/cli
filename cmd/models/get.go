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

	"github.com/openfga/fga-cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Read a Single Authorization Model",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		authorizationModelID := clientConfig.AuthorizationModelID
		options := client.ClientReadAuthorizationModelOptions{
			AuthorizationModelId: openfga.PtrString(authorizationModelID),
		}
		model, err := fgaClient.ReadAuthorizationModel(context.Background()).Options(options).Execute()
		if err != nil {
			fmt.Printf("Failed to get model %v due to %v", clientConfig.AuthorizationModelID, err)
			os.Exit(1)
		}

		modelJSON, err := json.Marshal(model)
		if err != nil {
			fmt.Printf("Failed to get model due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(modelJSON))
	},
}

func init() {
	getCmd.Flags().String("model-id", "", "Authorization Model ID")
	err := getCmd.MarkFlagRequired("model-id")
	if err != nil { //nolint:wsl
		fmt.Print(err)
		os.Exit(1)
	}
}
