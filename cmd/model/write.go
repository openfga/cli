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
	"fmt"
	"os"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/flags"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/utils"
)

func Write(
	ctx context.Context,
	fgaClient client.SdkClient,
	inputModel authorizationmodel.AuthzModel,
) (*client.ClientWriteAuthorizationModelResponse, error) {
	body := client.ClientWriteAuthorizationModelRequest{
		SchemaVersion:   inputModel.GetSchemaVersion(),
		TypeDefinitions: inputModel.GetTypeDefinitions(),
		Conditions:      inputModel.GetConditions(),
	}

	model, err := fgaClient.WriteAuthorizationModel(ctx).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to write model due to %w", err)
	}

	return model, nil
}

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write Authorization Model",
	Long:  "Writes a new authorization model.",
	Example: `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file=model.json
fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file=fga.mod
fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"type_definitions":[{"type":"user"},{"type":"document","relations":{"can_view":{"this":{}}},"metadata":{"relations":{"can_view":{"directly_related_user_types":[{"type":"user"}]}}}}],"schema_version":"1.1"}' --format=json`, //nolint:lll
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		var inputModel string
		if err := authorizationmodel.ReadFromInputFileOrArg(
			cmd,
			args,
			"file",
			false,
			&inputModel,
			openfga.PtrString(""),
			&writeInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		debug, _ := cmd.Flags().GetBool("debug")

		authModel := authorizationmodel.AuthzModel{}

		err = authModel.ReadModelFromString(inputModel, writeInputFormat)
		if err != nil {
			return err //nolint:wrapcheck
		}

		ctx := utils.WithDebugContext(cmd.Context(), debug)

		response, err := Write(ctx, fgaClient, authModel)
		if err != nil {
			return err
		}

		return output.Display(*response)
	},
}

var writeInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	writeCmd.Flags().String("store-id", "", "Store ID")
	writeCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON or DSL format")
	writeCmd.Flags().Var(&writeInputFormat, "format", `Authorization model input format. Can be "fga", "json", or "modular"`) //nolint:lll

	if err := flags.SetFlagRequired(writeCmd, "store-id", "cmd/model/write", false); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
