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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Read a Single Authorization Model",
	Long:    "Read an authorization model, pass in an empty model ID to get latest.",
	Example: `fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		response, err := authorizationmodel.ReadFromStore(cmd.Context(), clientConfig, fgaClient)
		if err != nil {
			return err //nolint:wrapcheck
		}

		authModel := authorizationmodel.AuthzModel{}
		authModel.Set(*response.AuthorizationModel)

		fields, err := cmd.Flags().GetStringArray("field")
		if err != nil {
			return fmt.Errorf("failed to parse field array flag due to %w", err)
		}

		if getOutputFormat == authorizationmodel.ModelFormatJSON {
			return output.Display(authModel.DisplayAsJSON(fields))
		}

		dslModel, err := authModel.DisplayAsDSL(fields)
		if err != nil {
			return fmt.Errorf("failed to display model due to %w", err)
		}

		fmt.Printf("%v", *dslModel)

		return nil
	},
}

var getOutputFormat = authorizationmodel.ModelFormatFGA

func init() {
	getCmd.Flags().String("model-id", "", "Authorization Model ID")
	getCmd.Flags().String("store-id", "", "Store ID")
	getCmd.Flags().StringArray("field", []string{"model"}, "Fields to display, choices are: id, created_at and model") //nolint:lll
	getCmd.Flags().Var(&getOutputFormat, "format", `Authorization model output format. Can be "fga" or "json"`)

	if err := getCmd.MarkFlagRequired("store-id"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/get", err)
		os.Exit(1)
	}
}
