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
	"errors"
	"fmt"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/cmd/model"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

type CreateStoreAndModelResponse struct {
	Store client.ClientCreateStoreResponse              `json:"store"`
	Model *client.ClientWriteAuthorizationModelResponse `json:"model,omitempty"`
}

func create(
	ctx context.Context, fgaClient client.SdkClient, storeName string,
) (*client.ClientCreateStoreResponse, error) {
	body := client.ClientCreateStoreRequest{Name: storeName}

	store, err := fgaClient.CreateStore(ctx).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create store %v due to %w", storeName, err)
	}

	return store, nil
}

func CreateStoreWithModel(
	ctx context.Context,
	fgaClient client.SdkClient,
	storeName string,
	inputModel string,
	inputFormat authorizationmodel.ModelFormat,
) (*CreateStoreAndModelResponse, error) {
	response := CreateStoreAndModelResponse{}

	if storeName == "" {
		return nil, errors.New(`required flag(s) "name" not set`) //nolint:goerr113
	}

	createStoreResponse, err := create(ctx, fgaClient, storeName)
	if err != nil {
		return nil, err
	}

	response.Store = *createStoreResponse

	err = fgaClient.SetStoreId(response.Store.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to set store ID: %w", err)
	}

	if inputModel != "" {
		authModel := authorizationmodel.AuthzModel{}

		err = authModel.ReadModelFromString(inputModel, inputFormat)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		createAuthZModelResponse, err := model.Write(ctx, fgaClient, authModel)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		response.Model = createAuthZModelResponse
	}

	return &response, nil
}

// createCmd represents the store create command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Store",
	Long:  "Create an OpenFGA store.",
	Example: `fga store create --name "FGA Demo Store" 

To set the created store id as an environment variable that will be used by the CLI, you can use the following command:

export FGA_STORE_ID=$(fga store create --model Model.fga | jq -r .store.id)
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		storeName, _ := cmd.Flags().GetString("name")
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client: %w", err)
		}

		var inputModel string
		if err := authorizationmodel.ReadFromInputFileOrArg(
			cmd,
			args,
			"model",
			true,
			&inputModel,
			&storeName,
			&createModelInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		response, err := CreateStoreWithModel(cmd.Context(), fgaClient, storeName, inputModel, createModelInputFormat)
		if err != nil {
			return err
		}

		return output.Display(response)
	},
}

var createModelInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	createCmd.Flags().String("name", "", "Store Name")
	createCmd.Flags().String("model", "", "Authorization Model File Name")
	createCmd.Flags().Var(&createModelInputFormat, "format", `Authorization model input format. Can be "fga", "json" or "modular`) //nolint:lll
}
