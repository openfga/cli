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
	"path"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	openfga "github.com/openfga/go-sdk"
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

const (
	progressBarWidth         = 40
	progressBarSleepDelay    = 10 // time.Millisecond
	progressBarThrottleValue = 65
	progressBarUpdateDelay   = 5 * time.Millisecond
)

// createStore creates a new store with the given client configuration and store data.
func createStore(
	clientConfig *fga.ClientConfig,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	fileName string,
) (*CreateStoreAndModelResponse, error) {
	storeDataName := storeData.Name
	if storeDataName == "" {
		storeDataName = strings.TrimSuffix(path.Base(fileName), ".fga.yaml")
	}

	createStoreAndModelResponse, err := CreateStoreWithModel(*clientConfig, storeDataName, storeData.Model, format)
	if err != nil {
		return nil, err
	}

	clientConfig.StoreID = createStoreAndModelResponse.Store.Id

	return createStoreAndModelResponse, nil
}

// updateStore updates an existing store with the given client configuration, store data, and store ID.
func updateStore(
	clientConfig *fga.ClientConfig,
	fgaClient client.SdkClient,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	storeID string,
) (*CreateStoreAndModelResponse, error) {
	store, err := fgaClient.GetStore(context.Background()).Execute()

	response := &CreateStoreAndModelResponse{
		Store: client.ClientCreateStoreResponse{
			Id: storeID,
		},
	}

	if err != nil {
		response.Store.CreatedAt = store.GetCreatedAt()
		response.Store.Name = store.GetName()
		response.Store.UpdatedAt = store.GetUpdatedAt()
	}

	authModel := authorizationmodel.AuthzModel{}
	clientConfig.StoreID = storeID

	if err := authModel.ReadModelFromString(storeData.Model, format); err != nil {
		return nil, fmt.Errorf("failed to read model: %w", err)
	}

	modelWriteRes, err := model.Write(fgaClient, authModel)
	if err != nil {
		return nil, fmt.Errorf("failed to write model: %w", err)
	}

	response.Model = modelWriteRes

	return response, nil
}

// importStore imports store data, either creating a new store or updating an existing one.
func importStore(
	clientConfig *fga.ClientConfig,
	fgaClient client.SdkClient,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	storeID string,
	maxTuplesPerWrite, maxParallelRequests int32,
	fileName string,
) (*CreateStoreAndModelResponse, error) {
	response, err := createOrUpdateStore(clientConfig, fgaClient, storeData, format, storeID, fileName)
	if err != nil {
		return nil, err
	}

	if len(storeData.Tuples) == 0 {
		return response, nil
	}

	err = importTuples(fgaClient, storeData.Tuples, maxTuplesPerWrite, maxParallelRequests)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func createOrUpdateStore(
	clientConfig *fga.ClientConfig,
	fgaClient client.SdkClient,
	storeData *storetest.StoreData,
	format authorizationmodel.ModelFormat,
	storeID string,
	fileName string,
) (*CreateStoreAndModelResponse, error) {
	if storeID == "" {
		return createStore(clientConfig, storeData, format, fileName)
	}

	return updateStore(clientConfig, fgaClient, storeData, format, storeID)
}

func importTuples(
	fgaClient client.SdkClient,
	tuples []openfga.TupleKey,
	maxTuplesPerWrite, maxParallelRequests int32,
) error {
	bar := createProgressBar(len(tuples))

	for index := 0; index < len(tuples); index += int(maxTuplesPerWrite) {
		end := index + int(maxTuplesPerWrite)
		if end > len(tuples) {
			end = len(tuples)
		}

		writeRequest := client.ClientWriteRequest{
			Writes: tuples[index:end],
		}
		if _, err := tuple.ImportTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests); err != nil {
			return fmt.Errorf("failed to import tuples: %w", err)
		}

		if err := bar.Add(end - index); err != nil {
			return fmt.Errorf("failed to update progress bar: %w", err)
		}

		time.Sleep(progressBarUpdateDelay)
	}

	if err := bar.Finish(); err != nil {
		return fmt.Errorf("failed to finish progress bar: %w", err)
	}

	return nil
}

func createProgressBar(total int) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetDescription("Importing tuples"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(progressBarWidth),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionThrottle(progressBarThrottleValue*progressBarSleepDelay),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("tuples"),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

// importCmd represents the get command.
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import Store Data",
	Long:    `Import a store: updating the name, model and appending the global tuples`,
	Example: "fga store import --file=model.fga.yaml",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		storeID, err := cmd.Flags().GetString("store-id")
		if err != nil {
			return fmt.Errorf("failed to get store-id: %w", err)
		}

		maxTuplesPerWrite, err := cmd.Flags().GetInt32("max-tuples-per-write")
		if err != nil {
			return fmt.Errorf("failed to parse max tuples per write: %w", err)
		}

		maxParallelRequests, err := cmd.Flags().GetInt32("max-parallel-requests")
		if err != nil {
			return fmt.Errorf("failed to parse parallel requests: %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to get file name: %w", err)
		}

		format, storeData, err := storetest.ReadFromFile(fileName, path.Dir(fileName))
		if err != nil {
			return fmt.Errorf("failed to read from file: %w", err)
		}

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client: %w", err)
		}

		createStoreAndModelResponse, err := importStore(&clientConfig, fgaClient, storeData, format,
			storeID, maxTuplesPerWrite, maxParallelRequests, fileName)
		if err != nil {
			return fmt.Errorf("failed to import store: %w", err)
		}

		// Print the response using output.Display without printing <nil>
		if outputErr := output.Display(createStoreAndModelResponse); outputErr != nil {
			return fmt.Errorf("failed to display output: %w", outputErr)
		}

		return nil
	},
}

func init() {
	importCmd.Flags().String("file", "", "File Name. The file should have the store")
	importCmd.Flags().String("store-id", "", "Store ID")
	importCmd.Flags().Int32("max-tuples-per-write", tuple.MaxTuplesPerWrite, "Max tuples per write chunk.")
	importCmd.Flags().Int32("max-parallel-requests", tuple.MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll

	if err := importCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/write", err)
		os.Exit(1)
	}
}
