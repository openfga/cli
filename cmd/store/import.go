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
	"fmt"
	"os"
	"path"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/cmd/model"
	"github.com/openfga/cli/cmd/tuple"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
)

// importStoreIODependencies defines IO dependencies for importing store
type importStoreIODependencies struct {
	createStoreWithModel func(
		clientConfig fga.ClientConfig,
		storeName string,
		inputModel string,
		inputFormat authorizationmodel.ModelFormat,
	) (*CreateStoreAndModelResponse, error)
	importTuples func(
		fgaClient client.SdkClient,
		body client.ClientWriteRequest,
		maxTuplesPerWrite int,
		maxParallelRequests int,
	) (*tuple.ImportResponse, error)
	modelWrite func(
		fgaClient client.SdkClient,
		inputModel authorizationmodel.AuthzModel,
	) (*client.ClientWriteAuthorizationModelResponse, error)
}

type ImportStoreResponse struct {
	*CreateStoreAndModelResponse
	Tuple *tuple.ImportResponse `json:"tuple"`
}

func importStore(
	clientConfig fga.ClientConfig,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	storeID string,
	maxTuplesPerWrite int,
	maxParallelRequests int,
	ioAggregator importStoreIODependencies,
) (*ImportStoreResponse, error) {
	var err error

	var fgaClient client.SdkClient

	response := &ImportStoreResponse{
		CreateStoreAndModelResponse: &CreateStoreAndModelResponse{},
	}
	if storeID == "" { //nolint:wsl
		createStoreAndModelResponse, err := ioAggregator.createStoreWithModel(
			clientConfig,
			storeData.Name,
			storeData.Model,
			format,
		)
		response.CreateStoreAndModelResponse = createStoreAndModelResponse
		if err != nil { //nolint:wsl
			return nil, err
		}
		clientConfig.StoreID = createStoreAndModelResponse.Store.Id //nolint:wsl
	} else {
		authModel := authorizationmodel.AuthzModel{}
		clientConfig.StoreID = storeID

		err = authModel.ReadModelFromString(storeData.Model, format)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		fgaClient, err = clientConfig.GetFgaClient()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		authorizationModelResponse, err := ioAggregator.modelWrite(fgaClient, authModel)
		if err != nil {
			return nil, fmt.Errorf("failed to write model due to %w", err)
		}

		response.Model = authorizationModelResponse
	}

	fgaClient, err = clientConfig.GetFgaClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FGA Client due to %w", err)
	}

	writeRequest := client.ClientWriteRequest{
		Writes: storeData.Tuples,
	}

	importTupleResponse, err := ioAggregator.importTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests)
	if err != nil {
		return nil, err
	}

	response.Tuple = importTupleResponse

	return response, nil
}

// importCmd represents the get command.
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import Store Data",
	Long:    `Import a store: updating the name, model and appending the global tuples`,
	Example: "fga store import --file=model.fga.yaml",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var createStoreAndModelResponse *ImportStoreResponse
		clientConfig := cmdutils.GetClientConfig(cmd)

		storeID, err := cmd.Flags().GetString("store-id")
		if err != nil {
			return fmt.Errorf("failed to get store-id %w", err)
		}

		maxTuplesPerWrite, err := cmd.Flags().GetInt("max-tuples-per-write")
		if err != nil {
			return fmt.Errorf("failed to parse max tuples per write due to %w", err)
		}

		maxParallelRequests, err := cmd.Flags().GetInt("max-parallel-requests")
		if err != nil {
			return fmt.Errorf("failed to parse parallel requests due to %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return err //nolint:wrapcheck
		}

		format, storeData, err := storetest.ReadFromFile(fileName, path.Dir(fileName))
		if err != nil {
			return err //nolint:wrapcheck
		}

		ioAggregator := importStoreIODependencies{
			createStoreWithModel: CreateStoreWithModel,
			importTuples:         tuple.ImportTuples,
			modelWrite:           model.Write,
		}
		createStoreAndModelResponse, err = importStore(clientConfig, storeData, format,
			storeID, maxTuplesPerWrite, maxParallelRequests, ioAggregator)
		if err != nil {
			return err
		}

		return output.Display(createStoreAndModelResponse)
	},
}

func init() {
	importCmd.Flags().String("file", "", "File Name. The file should have the store")
	importCmd.Flags().String("store-id", "", "Store ID")
	importCmd.Flags().Int("max-tuples-per-write", tuple.MaxTuplesPerWrite, "Max tuples per write chunk.")
	importCmd.Flags().Int("max-parallel-requests", tuple.MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll

	if err := importCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/write", err)
		os.Exit(1)
	}
}
