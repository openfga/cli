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

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func buildStoreData(config fga.ClientConfig, fgaClient client.SdkClient) (*storetest.StoreData, error) {
	// get the store
	store, err := fgaClient.GetStore(context.Background()).Execute()

	if err != nil {
		return nil, fmt.Errorf("unable to fetch store: %w", err)
	}

	model, err := authorizationmodel.ReadFromStore(config, fgaClient)

	if err != nil {
		return nil, err
	}

	authModel := authorizationmodel.AuthzModel{}
	authModel.Set(*model.AuthorizationModel)
	dsl, err := authModel.DisplayAsDSL([]string{"model"})

	if err != nil {
		return nil, fmt.Errorf("unable to get model dsl: %w", err)
	}

	storeData := &storetest.StoreData{
		Name:  store.Name,
		Model: *dsl,
	}

	return storeData, nil
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export store data",
	Long:    `Export a store to the export file format`,
	Example: "fga store export",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()

		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		storeData, err := buildStoreData(clientConfig, fgaClient)

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

			err = os.WriteFile(fileName, storeYaml, 0666)

			if err != nil {
				return err
			}
		}

		return output.Display(output.EmptyStruct{})
	},
}

func init() {
	fmt.Println("Initting exportCmd")

	exportCmd.Flags().String("output-file", "", "name of the file to export the store to")
	exportCmd.Flags().String("store-id", "", "store ID")
	exportCmd.Flags().String("model-id", "", "Authorization Model ID")

	err := exportCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
