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
	"github.com/openfga/cli/lib/output"
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
	Use:   "write",
	Short: "Write Authorization Model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		response, err := write(fgaClient, args[0])
		if err != nil {
			return err
		}

		return output.Display(cmd, *response) //nolint:wrapcheck
	},
}

func init() {
	writeCmd.Flags().String("store-id", "", "Store ID")

	if err := writeCmd.MarkFlagRequired("store-id"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/write", err)
		os.Exit(1)
	}
}
