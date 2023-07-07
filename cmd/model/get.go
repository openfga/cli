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
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/cli/lib/fga"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func getModel(clientConfig fga.ClientConfig, fgaClient client.SdkClient) (string, error) {
	authorizationModelID := clientConfig.AuthorizationModelID

	var err error

	var model *openfga.ReadAuthorizationModelResponse

	if authorizationModelID != "" {
		options := client.ClientReadAuthorizationModelOptions{
			AuthorizationModelId: openfga.PtrString(authorizationModelID),
		}
		model, err = fgaClient.ReadAuthorizationModel(context.Background()).Options(options).Execute()
	} else {
		options := client.ClientReadLatestAuthorizationModelOptions{}
		model, err = fgaClient.ReadLatestAuthorizationModel(context.Background()).Options(options).Execute()
	}

	if err != nil {
		return "", fmt.Errorf("failed to get model %v due to %w", clientConfig.AuthorizationModelID, err)
	}

	modelJSON, err := json.Marshal(model)
	if err != nil {
		return "", fmt.Errorf("failed to get model due to %w", err)
	}

	return string(modelJSON), nil
}

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:              "get",
	Short:            "Read a Single Authorization Model",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}
		output, err := getModel(clientConfig, fgaClient)
		if err != nil {
			return err
		}

		fmt.Print(output)

		return nil
	},
}

func init() {
	getCmd.Flags().String("model-id", "", "Authorization Model ID")
	getCmd.Flags().String("store-id", "", "Store ID")

	if err := getCmd.MarkFlagRequired("store-id"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/get", err)
		os.Exit(1)
	}
}
