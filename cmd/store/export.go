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
	"math"
	"os"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/confirmation"
	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/cli/internal/tuple"
)

// defaultTuplePageSize defines the default number of pages to return when calling Read.
const (
	defaultMaxTupleCount = 100
)

// buildStoreData compiles all the data necessary to output to file or stdout,
// or returns an error if this was not successful.
func buildStoreData(config fga.ClientConfig, fgaClient client.SdkClient, maxTupleCount uint) (*storetest.StoreData, error) { //nolint:lll
	// get the store
	store, _ := fgaClient.GetStore(context.Background()).Execute()

	var storeName string
	if store != nil {
		storeName = store.Name
	}

	// get the model
	model, err := authorizationmodel.ReadFromStore(config, fgaClient)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	authModel := authorizationmodel.AuthzModel{}
	authModel.Set(*model.AuthorizationModel)

	dsl, err := authModel.DisplayAsDSL([]string{"model"})
	if err != nil {
		return nil, fmt.Errorf("unable to get model dsl: %w", err)
	}

	// get the tuples
	maxPages := int(math.Ceil(float64(maxTupleCount) / float64(tuple.DefaultReadPageSize)))

	rawTuples, err := tuple.Read(fgaClient, &client.ClientReadRequest{}, maxPages)
	if err != nil {
		return nil, fmt.Errorf("unable to read tuples: %w", err)
	}

	tuples := rawTuples.GetTuples()
	maxTuplesInOutput := int(math.Min(float64(len(tuples)), float64(maxTupleCount)))
	outputTuples := make([]client.ClientContextualTupleKey, 0, maxTuplesInOutput)

	for _, t := range tuples[:maxTuplesInOutput] {
		outputTuples = append(outputTuples, t.GetKey())
	}

	// get the assertions
	assertionOptions := client.ClientReadAssertionsOptions{
		AuthorizationModelId: authModel.ID,
	}

	assertionResponse, err := fgaClient.ReadAssertions(context.Background()).Options(assertionOptions).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to read assertions: %w", err)
	}

	assertions := assertionResponse.GetAssertions()
	modelChecks := map[string]storetest.ModelTestCheck{}

	for _, assertion := range assertions {
		key := fmt.Sprintf("%s|%s", assertion.TupleKey.User, assertion.TupleKey.Object)
		_, exists := modelChecks[key]

		if !exists {
			modelChecks[key] = storetest.ModelTestCheck{
				User:       assertion.TupleKey.User,
				Object:     assertion.TupleKey.Object,
				Assertions: map[string]bool{},
			}
		}

		modelChecks[key].Assertions[assertion.GetTupleKey().Relation] = assertion.Expectation
	}

	checks := []storetest.ModelTestCheck{}
	for _, value := range modelChecks {
		checks = append(checks, value)
	}

	storeData := &storetest.StoreData{
		Name:   storeName,
		Model:  *dsl,
		Tuples: outputTuples,
		Tests: []storetest.ModelTest{
			{
				Name:  "Tests",
				Check: checks,
			},
		},
	}

	return storeData, nil
}

// exportCmd represents the export store command.
var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export store data",
	Long:    `Export a store to YAML`,
	Example: "fga store export",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		maxTupleCount, _ := cmd.Flags().GetUint("max-tuples")
		storeData, err := buildStoreData(clientConfig, fgaClient, maxTupleCount)
		if err != nil {
			return fmt.Errorf("failed to export store: %w", err)
		}

		if storeData != nil {
			storeYaml, err := yaml.Marshal(storeData)
			if err != nil {
				return fmt.Errorf("unable to marshal storedata yaml: %w", err)
			}

			fileName, _ := cmd.Flags().GetString("output-file")
			if fileName == "" {
				fmt.Println(string(storeYaml))

				return nil
			}

			if _, err := os.Stat(fileName); err == nil {
				confirm, err := confirmation.AskForConfirmation("File exists, overwrite?")
				if err != nil {
					return fmt.Errorf("prompt failed due to %w", err)
				}

				if !confirm {
					fmt.Println("cancelled")

					return output.Display(output.EmptyStruct{})
				}
			}

			err = os.WriteFile(fileName, storeYaml, 0o600) //nolint:gomnd
			if err != nil {
				return err //nolint:wrapcheck
			}

			fmt.Printf("model written to %s\n", fileName)
		}

		return output.Display(output.EmptyStruct{})
	},
}

func init() {
	exportCmd.Flags().String("output-file", "", "name of the file to export the store to")
	exportCmd.Flags().String("store-id", "", "store ID")
	exportCmd.Flags().String("model-id", "", "Authorization Model ID")
	exportCmd.Flags().Uint("max-tuples", defaultMaxTupleCount, "max number of tuples to return in the output")

	err := exportCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
