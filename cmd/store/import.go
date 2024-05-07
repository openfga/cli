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
	"strings"

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

func importStore(
	clientConfig fga.ClientConfig,
	fgaClient client.SdkClient,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	storeID string,
	maxTuplesPerWrite int,
	maxParallelRequests int,
	fileName string,
) (*CreateStoreAndModelResponse, error) {
	var err error
	var response *CreateStoreAndModelResponse //nolint:wsl
	if storeID == "" {                        //nolint:wsl,nestif
		storeDataName := storeData.Name
		if storeDataName == "" {
			storeDataName = strings.TrimSuffix(path.Base(fileName), ".fga.yaml")
		}
		createStoreAndModelResponse, err := CreateStoreWithModel(clientConfig, storeDataName, //nolint:wsl
			storeData.Model, format)
		response = createStoreAndModelResponse
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

		_, err := model.Write(fgaClient, authModel)
		if err != nil {
			return nil, fmt.Errorf("failed to write model due to %w", err)
		}
	}

	fgaClient, err = clientConfig.GetFgaClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FGA Client due to %w", err)
	}

	writeRequest := client.ClientWriteRequest{
		Writes: storeData.Tuples,
	}

	_, err = tuple.ImportTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return response, nil
}

// importCmd represents the get command.
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import Store Data",
	Long:    `Import a store: updating the name, model and appending the global tuples`,
	Example: "fga store import --file=model.fga.yaml",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var createStoreAndModelResponse *CreateStoreAndModelResponse
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

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		createStoreAndModelResponse, err = importStore(clientConfig, fgaClient, storeData, format,
			storeID, maxTuplesPerWrite, maxParallelRequests, fileName)
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
