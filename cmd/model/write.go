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

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func write(fgaClient client.SdkClient, text string) (*client.ClientWriteAuthorizationModelResponse, error) {
	body := &client.ClientWriteAuthorizationModelRequest{}

	err := json.Unmarshal([]byte(text), &body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model due to %w", err)
	}

	model, err := fgaClient.WriteAuthorizationModel(context.Background()).Body(*body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to write model due to %w", err)
	}

	return model, nil
}

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:     "write",
	Short:   "Write Authorization Model",
	Args:    cobra.MaximumNArgs(1),
	Example: `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file=model.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to parse file name due to %w", err)
		}

		var inputModel string
		if fileName != "" {
			file, err := os.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("failed to read file %s due to %w", fileName, err)
			}
			inputModel = string(file)
		} else {
			if len(args) == 0 || args[0] == "-" {
				return cmd.Help() //nolint:wrapcheck
			}
			inputModel = args[0]
		}

		response, err := write(fgaClient, inputModel)
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	writeCmd.Flags().String("store-id", "", "Store ID")
	writeCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON format")

	if err := writeCmd.MarkFlagRequired("store-id"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/write", err)
		os.Exit(1)
	}
}
